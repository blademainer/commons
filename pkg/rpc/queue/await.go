package queue

import (
	"fmt"
	"github.com/blademainer/commons/pkg/logger"
	mqttpb "github.com/blademainer/commons/pkg/rpc/queue/proto"
	"sort"
	"sync"
	"time"
)

type awaitKeeper struct {
	sync.RWMutex

	awaitCh       chan *awaitEntry // messageId -> callbackFunc
	deleteEntryCh chan *awaitEntry
	messageIdMap  map[string]*awaitEntry
	responseCh    *chan *mqttpb.QueueMessage // chan for responses
	ttlEntries    []*awaitEntry
	tickInterval  time.Duration
}

type awaitEntry struct {
	messageId    string
	awaitTimeout time.Duration
	messageCh    chan *mqttpb.QueueMessage
	errorCh      chan error
	ttl          time.Time
}

func (entry *awaitEntry) Close() {
	close(entry.messageCh)
	close(entry.errorCh)
}

func (keeper *awaitKeeper) Len() int {
	return len(keeper.ttlEntries)
}

func (keeper *awaitKeeper) Less(i, j int) bool {
	return keeper.ttlEntries[i].ttl.UnixNano() < keeper.ttlEntries[j].ttl.UnixNano()
}

func (keeper *awaitKeeper) Swap(i, j int) {
	keeper.ttlEntries[i], keeper.ttlEntries[j] = keeper.ttlEntries[j], keeper.ttlEntries[i]
}

func newAwaitKeeper(opts *Options, responseMessageCh *chan *mqttpb.QueueMessage) *awaitKeeper {
	keeper := &awaitKeeper{}
	keeper.awaitCh = make(chan *awaitEntry, opts.awaitQueueSize)
	keeper.deleteEntryCh = make(chan *awaitEntry, opts.awaitQueueSize)
	keeper.messageIdMap = make(map[string]*awaitEntry)
	keeper.ttlEntries = make([]*awaitEntry, 0)
	keeper.tickInterval = opts.tickInterval
	keeper.responseCh = responseMessageCh
	return keeper
}

func (s *defaultServer) startKeeper() {
	s.keeper.startLoop(s.doneCh)
	s.keeper.startTick(s.doneCh)
}

func (keeper *awaitKeeper) startLoop(doneChan chan struct{}) {
	go func() {
		for {
			select {
			case _, closed := <-doneChan: // global closed
				if closed {
					logger.Warnf("func: startKeeper stopping...")
					return
				}
			case entry := <-keeper.awaitCh: // watch entry
				keeper.insertTtlEntry(entry)
			case response := <-*keeper.responseCh: // new response
				if response == nil {
					continue
				}
				keeper.handleAwaitResponse(response)
			case toDelete := <-keeper.deleteEntryCh: // ttl entry
				keeper.handleTtlResponse(toDelete)
			}
		}
	}()
}

func (keeper *awaitKeeper) handleTtlResponse(entry *awaitEntry) {
	entry, found := keeper.messageIdMap[entry.messageId]
	if !found {
		return
	}
	defer delete(keeper.messageIdMap, entry.messageId)
	defer entry.Close()
	e := &TimeoutError{message: "timed out"}
	select {
	case entry.errorCh <- e:
	default:
		if logger.IsDebugEnabled() {
			logger.Debugf("failed to push to entry's error chan. maybe it's been removed by ttl ticker, entry is: %v", entry)
		}
	}

}

func (keeper *awaitKeeper) handleAwaitResponse(message *mqttpb.QueueMessage) {
	entry, found := keeper.messageIdMap[message.MessageId]
	if !found {
		return
	}
	keeper.Lock()
	defer keeper.Unlock()
	delete(keeper.messageIdMap, entry.messageId)
	defer entry.Close()
	select {
	case entry.messageCh <- message:
	default:
		logger.Infof("failed to push to chan, message: %v entry: %v", message, entry)
	}
}

//func (keeper *awaitKeeper) deleteTtlEntry(entry *awaitEntry) {
//	delete(keeper.messageIdMap, entry.messageId)
//	for i, e := range keeper.ttlEntries {
//		if e.messageId == entry.messageId {
//			keeper.ttlEntries[i] = nil
//			deleteElement(keeper.ttlEntries, i)
//		}
//	}
//}
//
//func deleteElement(entries []*awaitEntry, index int) {
//	i := 2
//
//	// Remove the element at index i from a.
//	entries[i] = entries[len(entries)-1] // Copy last element to index i.
//	entries[len(entries)-1] = nil        // Erase last element (write zero value).
//	entries = entries[:len(entries)-1]   // Truncate slice.
//}

func (keeper *awaitKeeper) insertTtlEntry(entry *awaitEntry) {
	keeper.Lock()
	defer keeper.Unlock()
	keeper.messageIdMap[entry.messageId] = entry
	keeper.ttlEntries = append(keeper.ttlEntries, entry)
	sort.Sort(keeper)
}

func (s *defaultServer) watchMessageId(messageId string, await time.Duration, messageCh chan *mqttpb.QueueMessage, errorCh chan error) error {
	entry := &awaitEntry{}
	entry.messageId = messageId
	entry.awaitTimeout = await
	entry.messageCh = messageCh
	entry.errorCh = errorCh
	entry.ttl = time.Unix(0, time.Now().UnixNano()+await.Nanoseconds())
	select {
	case s.keeper.awaitCh <- entry:
		return nil
	default:
		e := &QueueFulledError{fmt.Sprintf("queue failed, size: %v", len(s.keeper.awaitCh))}
		return e
	}
}

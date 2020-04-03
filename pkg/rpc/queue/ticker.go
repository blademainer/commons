package queue

import (
	"github.com/blademainer/commons/pkg/logger"
	"sort"
	"time"
)

func (keeper *awaitKeeper) startTick(doneChan chan struct{}) {
	tick := time.Tick(keeper.tickInterval)
	go func() {
		for {
			select {
			case now := <-tick:
				keeper.doTick(now)
			case <-doneChan:
				logger.Infof("stopping tick")
				return
			}
		}
	}()
}

func (keeper *awaitKeeper) doTick(now time.Time) {
	if logger.IsDebugEnabled() {
		logger.Debugf("startTick...%v", now.Format(time.RFC3339Nano))
		logger.Debugf("entries: %v", keeper.ttlEntries)
	}

	subIndex := keeper.findBestPos(now)

	//for i, e := range keeper.ttlEntries {
	//	if e.ttl.Before(now) {
	//		subIndex = i + 1
	//	}
	//}

	subset := keeper.subset(subIndex)
	if subset == nil {
		if logger.IsDebugEnabled() {
			logger.Debugf("null subset. subIndex: %v entries: %v", subIndex, keeper.ttlEntries)
		}
		return
	}
	if logger.IsDebugEnabled() {
		logger.Debugf("found ttl subset size: %v, subIndex: %v, least entry size: %v", len(subset), subIndex, len(keeper.ttlEntries))
	}

	// delete ttl entry
	for _, s := range subset {
		keeper.deleteEntryCh <- s
	}
}

func (keeper *awaitKeeper) findBestPos(now time.Time) int {
	keeper.RLock()
	defer keeper.RUnlock()
	length := len(keeper.ttlEntries)
	nano := now.UnixNano()
	result := sort.Search(length, func(i int) bool {
		return keeper.ttlEntries[i].ttl.UnixNano() >= nano
	})
	return result
}

func (keeper *awaitKeeper) subset(subIndex int) []*awaitEntry {
	if subIndex <= 0 || subIndex > len(keeper.ttlEntries) {
		return nil
	}
	keeper.Lock()
	defer keeper.Unlock()
	if subIndex == len(keeper.ttlEntries) {
		subset := keeper.ttlEntries
		keeper.ttlEntries = make([]*awaitEntry, 0)
		return subset
	}
	subset := keeper.ttlEntries[:subIndex]
	keeper.ttlEntries = keeper.ttlEntries[subIndex:]
	return subset
}

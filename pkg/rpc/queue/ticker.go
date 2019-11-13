package queue

import (
	"fmt"
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
		logger.Debugf("found ttl subset size: %v, subIndex: %v, entries: %v", len(subset), subIndex, keeper.ttlEntries)
	}

	// delete ttl entry
	for _, s := range subset {
		keeper.handleTtlResponse(s)
	}
}

func (keeper *awaitKeeper) findBestPos(now time.Time) int {
	keeper.RLock()
	defer keeper.RUnlock()
	length := len(keeper.ttlEntries)
	i := sort.Search(length, func(i int) bool {
		if i == length-1 && keeper.ttlEntries[i].ttl.Before(now) {
			return true
		}
		return (keeper.ttlEntries[i].ttl.Before(now) || keeper.ttlEntries[i].ttl.Equal(now)) && keeper.ttlEntries[i+1].ttl.After(now)
	})
	return i
}

func foundBestPos(arr []int, search int) int {
	right := len(arr) - 1
	mid := right / 2
	left := 0
	for mid >= left && mid < right {
		if search == arr[mid] || (arr[mid] <= search && search < arr[mid+1]) {
			return mid
		} else if search < arr[mid] {
			mid = (mid + left) / 2
		} else {
			mid = (mid + right) / 2
		}
		fmt.Println(mid)
		if mid == 0 {
			if arr[mid] == search {
				return 0
			} else {
				return -1
			}
		}
	}
	return -1
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

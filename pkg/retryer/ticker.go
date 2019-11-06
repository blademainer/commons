package retryer

import (
	"github.com/blademainer/commons/pkg/logger"
	"time"
)

func (d *defaultRetryer) tick() {
	tick := time.Tick(d.Interval)
	go func() {
		for {
			select {
			case now := <-tick:
				d.doTick(now)
			case <-d.doneChan:
				return
			}
		}
	}()
}

func (d *defaultRetryer) doTick(now time.Time) {
	if logger.IsDebugEnabled() {
		logger.Debugf("tick...%v", now.Format(time.RFC3339Nano))
		logger.Debugf("entries: %v", d.retryEntries)
	}

	subIndex := -1
	for i, e := range d.retryEntries {
		if e.nextInvokeTime.Before(now) {
			subIndex = i + 1
		}
	}

	subset := d.subset(subIndex)
	if subset == nil {
		if logger.IsDebugEnabled() {
			logger.Debugf("found null subset! subIndex: %v entries: %v", subIndex, d.retryEntries)
		}
		return
	}
	if logger.IsDebugEnabled() {
		logger.Debugf("found subset size: %v, subIndex: %v, entries: %v", len(subset), subIndex, d.retryEntries)
	}

	for _, e := range subset {
		go d.doRetry(e)
	}
}

func (d *defaultRetryer) subset(subIndex int) []*retryEntry {
	if subIndex <= 0 || subIndex > len(d.retryEntries) {
		return nil
	}
	if subIndex == len(d.retryEntries) {
		return d.retryEntries
	}
	d.Lock()
	defer d.Unlock()
	subset := d.retryEntries[:subIndex]
	d.retryEntries = d.retryEntries[subIndex:]
	return subset
}

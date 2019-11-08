package retryer

import (
	"github.com/blademainer/commons/pkg/logger"
	recover2 "github.com/blademainer/commons/pkg/recover"
	"time"
)

func (d *defaultRetryer) tick() {
	tick := time.Tick(d.tickInterval)
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
	d.Lock()
	defer d.Unlock()
	if subIndex == len(d.retryEntries) {
		subset := d.retryEntries
		d.retryEntries = make([]*retryEntry, 0)
		return subset
	}
	subset := d.retryEntries[:subIndex]
	d.retryEntries = d.retryEntries[subIndex:]
	return subset
}

func (d *defaultRetryer) doRetry(entry *retryEntry) {
	defer recover2.Recover()
	e := d.invoke(entry.fn)

	if entry.retryTimes >= d.maxRetryTimes {
		e = &LimitedError{innerError: e}
		return
	}

	defer d.reportEvent(entry, e)
	if !IsRetryError(e) {
		return
	}
	d.afterFail(entry)
}

func (d *defaultRetryer) reportEvent(entry *retryEntry, err error) {
	event := RetryEvent{}
	event.RetryTimes = entry.retryTimes
	event.Time = time.Now()
	event.Fn = entry.fn
	event.Error = err
	if err == nil || !IsRetryError(err) {
		event.Success = true
	} else {
		event.Success = false
	}

	select {
	case d.retryEventChan <- event:
		if logger.IsDebugEnabled() {
			logger.Debugf("report event: %v", event)
		}
	default:
		if logger.IsDebugEnabled() {
			logger.Debugf("event chan is full, discard event: %v", event)
		}
	}
}

func (d *defaultRetryer) afterFail(entry *retryEntry) {
	nextRetryTime := d.nextRetryTime(entry.retryTimes)
	if logger.IsDebugEnabled() {
		logger.Debugf("entry: %v next retry time: %v", entry, nextRetryTime.Format(time.RFC3339Nano))
	}
	newE := &retryEntry{
		fn:             entry.fn,
		retryTimes:     entry.retryTimes + 1,
		nextInvokeTime: nextRetryTime,
	}
	d.timeoutFn(newE)
}

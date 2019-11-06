package retryer

import (
	"context"
	"errors"
	"fmt"
	"github.com/blademainer/commons/pkg/logger"
	recover2 "github.com/blademainer/commons/pkg/recover"
	"math/big"
	"sort"
	"time"
)

func (d *defaultRetryer) invoke(do Do) error {
	timeout, cancelFunc := context.WithTimeout(context.TODO(), d.Timeout)
	defer cancelFunc()
	var err error
	doneCh := make(chan struct{}, 1)
	go func() {
		defer recover2.WithRecover(func() {
			doneCh <- struct{}{}
		})
		err = do(timeout)
	}()
	done := timeout.Done()
	for {
		select {
		case <-done:
			err := fmt.Errorf("context deadline exceeded and we should retry later")
			e := &RetryError{innerError: err}
			return e
		case <-doneCh:
			// done
			return err
		case <-d.doneChan:
			err := errors.New("stopped")
			return err
		}
	}

}

func (d *defaultRetryer) getRetryEntry(do Do, retryTimes int) *retryEntry {
	entry := &retryEntry{}
	entry.fn = do
	seconds := nextRetryNanoSeconds(time.Now(), d.Interval, retryTimes, d.GrowthRate)
	entry.nextInvokeTime = time.Unix(0, seconds)
	return entry
}

// now + retryTimes*interval*growth
func nextRetryNanoSeconds(now time.Time, interval time.Duration, retryTimes int, growth float32) int64 {
	resultFloat := big.NewFloat(0)
	nowFloat := big.NewFloat(float64(now.UnixNano()))
	glowthFloat := big.NewFloat(float64(growth))
	intervalFloat := big.NewFloat(float64(interval.Nanoseconds()))
	retryTimesFloat := big.NewFloat(float64(retryTimes + 1))

	resultFloat = resultFloat.Mul(retryTimesFloat, intervalFloat)
	resultFloat = resultFloat.Mul(resultFloat, glowthFloat)
	resultFloat = resultFloat.Add(resultFloat, nowFloat)
	i, _ := resultFloat.Int64()
	//fmt.Println(accuracy)
	return i
}

// insert to retryChan
func (d *defaultRetryer) timeoutFn(entry *retryEntry) {
	select {
	case d.retryChan <- entry:
		if logger.IsDebugEnabled() {
			logger.Debugf("insert entry: %v", entry)
		}
	default:
		// chan is fulls
		if logger.IsDebugEnabled() {
			logger.Debugf("retry chan is full, now entry: %v", entry)
		}

		d.discard(entry)
	}
}

// consume from retryChan
func (d *defaultRetryer) consumeRetryChan() {
	go func() {
		for {
			select {
			case entry := <-d.retryChan:
				d.insertRetryEntry(entry)
			case <-d.doneChan:
				logger.Warnf("stopped HandleRetryChan by doneChan.")
				return
			}
		}
	}()
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

func (d *defaultRetryer) doRetry(entry *retryEntry) {
	defer recover2.Recover()
	e := d.invoke(entry.fn)

	if entry.retryTimes >= d.MaxRetryTimes {
		e = &LimitedError{innerError: e}
		return
	}

	defer d.reportEvent(entry, e)
	if !IsRetryError(e) {
		return
	}
	d.afterFail(entry)
}

func (d *defaultRetryer) afterFail(entry *retryEntry) {

	nextRetryNanoSeconds := nextRetryNanoSeconds(time.Now(), d.Interval, entry.retryTimes, d.GrowthRate)
	nextRetryTime := time.Unix(0, nextRetryNanoSeconds)
	newE := &retryEntry{
		fn:             entry.fn,
		retryTimes:     entry.retryTimes + 1,
		nextInvokeTime: nextRetryTime,
	}
	d.timeoutFn(newE)
}

func (d *defaultRetryer) discard(entry *retryEntry) {
	d.Lock()
	defer d.Unlock()
	if entry == nil || len(d.retryEntries) == 0 {
		return
	}
	switch d.DiscardStrategy {
	case DiscardStrategyRejectNew:
		if logger.IsDebugEnabled() {
			logger.Debugf("discard new entry: %v by strategy: DiscardStrategyRejectNew", entry)
		}
		return
	case DiscardStrategyLatest:
		if logger.IsDebugEnabled() {
			logger.Debugf("discard latest entry: %v by strategy: DiscardStrategyLatest", d.retryEntries[1:1])
		}
		d.retryEntries = d.retryEntries[1:]
	case DiscardStrategyEarliest:
		if logger.IsDebugEnabled() {
			logger.Debugf("discard earliest entry: %v by strategy: DiscardStrategyEarliest", d.retryEntries[len(d.retryEntries)-1:len(d.retryEntries)])
		}
		d.retryEntries = d.retryEntries[:len(d.retryEntries)-1]
	}
	d.Unlock()
	defer d.Lock()
	d.insertRetryEntry(entry)
}

func (d *defaultRetryer) insertRetryEntry(entry *retryEntry) {
	d.Lock()
	defer d.Unlock()
	d.retryEntries = append(d.retryEntries, entry)
	sort.Sort(d)
}

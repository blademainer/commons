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

// now + retryTimes*interval*growth
func nextRetryNanoSeconds(now time.Time, interval time.Duration, retryTimes int, growth int) int64 {
	resultFloat := big.NewFloat(float64(nextRetryDelayNanoseconds(interval, retryTimes, growth)))
	nowFloat := big.NewFloat(float64(now.UnixNano()))
	resultFloat = resultFloat.Add(resultFloat, nowFloat)
	i, _ := resultFloat.Int64()
	//fmt.Println(accuracy)
	return i
}

// retryTimes*interval*growth
func nextRetryDelayNanoseconds(interval time.Duration, retryTimes int, growth int) int64 {
	resultFloat := big.NewInt(0)
	growthFloat := big.NewInt(int64(growth))
	intervalFloat := big.NewInt(interval.Nanoseconds())
	retryTimesFloat := big.NewInt(int64(retryTimes + 1))

	resultFloat = resultFloat.Exp(growthFloat, retryTimesFloat, nil)
	resultFloat = resultFloat.Mul(resultFloat, intervalFloat)
	i := resultFloat.Int64()
	//fmt.Println(accuracy)
	return i
}

func (d *defaultRetryer) nextRetryTime(retryTimes int) time.Time {
	nanoSeconds := d.nextRetryNanoSeconds(retryTimes)
	return time.Unix(0, nanoSeconds)
}

func (d *defaultRetryer) nextRetryNanoSeconds(retryTimes int) int64 {
	nextRetryNanoSeconds := d.calculator.NextRetryDelayNanoseconds(d.tickInterval, retryTimes)
	return time.Now().UnixNano() + nextRetryNanoSeconds
}

func (d *defaultRetryer) invoke(do Do) error {
	timeout, cancelFunc := context.WithTimeout(context.TODO(), d.timeout)
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
			e := &RetryError{InnerError: err}
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
	nanoseconds := d.calculator.NextRetryDelayNanoseconds(d.tickInterval, retryTimes)
	seconds := time.Now().UnixNano() + nanoseconds
	entry.nextInvokeTime = time.Unix(0, seconds)
	return entry
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

func (d *defaultRetryer) discard(entry *retryEntry) {
	d.Lock()
	defer d.Unlock()
	if entry == nil || len(d.retryEntries) == 0 {
		return
	}
	switch d.discardStrategy {
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

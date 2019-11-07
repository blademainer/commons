package retryer

import (
	"context"
	"fmt"

	"sync"
	"time"
)

// Strategies when retry queue is full
type DiscardStrategy int

const (
	DiscardStrategyEarliest  DiscardStrategy = iota + 1 // discard the head of retry queue
	DiscardStrategyLatest                               // discard the last element in retry queue
	DiscardStrategyRejectNew                            // reject new retry job
)

type Do func(ctx context.Context) error

type Retryer interface {
	SetRetryTimeCalculator(calculator RetryTimeCalculator) error
	Invoke(do Do) error
	GetEvent() chan RetryEvent
	Stop() error
}

type RetryEvent struct {
	Fn         Do
	Time       time.Time
	RetryTimes int
	Success    bool
	Error      error
}

func (r RetryEvent) String() string {
	return fmt.Sprintf("RetryEvent[Do: %v, Time: %v, RetryTimes: %v, Success: %v, Error: %v]", r.Fn, r.Time, r.RetryTimes, r.Success, r.Error)
}

type defaultRetryer struct {
	sync.RWMutex
	calculator RetryTimeCalculator

	maxRetryTimes   int
	retryChan       chan *retryEntry
	retryEntries    []*retryEntry
	retryEventChan  chan RetryEvent
	doneChan        chan struct{}
	tickInterval    time.Duration
	timeout         time.Duration
	discardStrategy DiscardStrategy
}

func (d *defaultRetryer) SetRetryTimeCalculator(calculator RetryTimeCalculator) error {
	d.calculator = calculator
	return nil
}

func (d *defaultRetryer) Stop() error {
	d.doneChan <- struct{}{}
	close(d.doneChan)
	close(d.retryChan)
	close(d.retryEventChan)
	return nil
}

func (d *defaultRetryer) GetEvent() chan RetryEvent {
	return d.retryEventChan
}

func (d *defaultRetryer) String() string {
	return fmt.Sprintf("RetryStrategy: %v, retryEntries: %v", d.calculator, d.retryEntries)
}

func (d *defaultRetryer) Len() int {
	return len(d.retryEntries)
}

func (d *defaultRetryer) Less(i, j int) bool {
	return d.retryEntries[i].nextInvokeTime.UnixNano() < d.retryEntries[j].nextInvokeTime.UnixNano()
}

func (d *defaultRetryer) Swap(i, j int) {
	d.retryEntries[i], d.retryEntries[j] = d.retryEntries[j], d.retryEntries[i]
}

type retryEntry struct {
	nextInvokeTime time.Time
	fn             Do
	retryTimes     int
}

func (r *retryEntry) String() string {
	return fmt.Sprintf("RetryEntry:[nextInvokeTime: %v, fn: %v, retryTimes: %v]", r.nextInvokeTime.Format(time.RFC3339Nano), r.fn, r.retryTimes)
}

func NewDoubleGrowthRetryer(timeout time.Duration) (Retryer, error) {
	strategy := NewDefaultDoubleGrowthRateRetryStrategy()
	return NewRetryer(strategy, 10, 100, timeout, timeout, DiscardStrategyEarliest)
}

func NewRetryer(calculator RetryTimeCalculator, maxRetryTimes int, maxRetryEntriesInQueue int, tickInterval time.Duration, timeout time.Duration, discardStrategy DiscardStrategy) (Retryer, error) {
	retryer := &defaultRetryer{}
	retryer.doneChan = make(chan struct{})
	err := retryer.SetRetryTimeCalculator(calculator)
	if err != nil {
		return nil, err
	}
	retryer.retryEventChan = make(chan RetryEvent, maxRetryEntriesInQueue)
	retryer.retryChan = make(chan *retryEntry, maxRetryEntriesInQueue)
	retryer.maxRetryTimes = maxRetryTimes
	retryer.tickInterval = tickInterval
	retryer.timeout = timeout
	retryer.discardStrategy = discardStrategy
	retryer.consumeRetryChan()
	retryer.tick()
	return retryer, nil
}

//func (d *defaultRetryer) SetRetryStrategy(strategy *RetryStrategy) error {
//	if strategy.GrowthRate < 0 {
//		return errors.New("growthRate must greater equals 0")
//	}
//	d.calculator = strategy
//	return nil
//}

func (d *defaultRetryer) Invoke(do Do) error {
	err := d.invoke(do)
	if IsRetryError(err) && d.maxRetryTimes > 0 {
		// first time of timeout
		entry := d.getRetryEntry(do, 0)
		d.timeoutFn(entry)
	}

	return err
}

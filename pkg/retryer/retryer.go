package retryer

import (
	"context"
	"errors"
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

type RetryStrategy struct {
	Timeout             time.Duration
	Interval            time.Duration
	GrowthRate          float32 // rate of growth retry delay.
	MaxRetrySizeInQueue uint16  // Max size of retry function in queue
	DiscardStrategy     DiscardStrategy
	MaxRetryTimes       int
}

func NewDefaultDoubleGrowthRateRetryStrategy() *RetryStrategy {
	return NewDoubleGrowthRateRetryStrategy(5*time.Second, 5*time.Second, 10, 1024, DiscardStrategyEarliest)
}

func NewDoubleGrowthRateRetryStrategy(timeout time.Duration, interval time.Duration, maxRetryTimes int, maxRetrySizeInQueue uint16, discardStrategy DiscardStrategy) *RetryStrategy {
	strategy := &RetryStrategy{}
	strategy.Timeout = timeout
	strategy.Interval = interval
	strategy.MaxRetrySizeInQueue = maxRetrySizeInQueue
	strategy.DiscardStrategy = discardStrategy
	strategy.MaxRetryTimes = maxRetryTimes
	strategy.GrowthRate = 2.0
	return strategy
}

type Retryer interface {
	SetRetryStrategy(strategy *RetryStrategy) error
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
	*RetryStrategy

	retryChan      chan *retryEntry
	retryEntries   []*retryEntry
	retryEventChan chan RetryEvent
	doneChan       chan struct{}
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
	return fmt.Sprintf("RetryStrategy: %v, retryEntries: %v", d.RetryStrategy, d.retryEntries)
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

func NewRetryer(strategy *RetryStrategy) (Retryer, error) {
	retryer := &defaultRetryer{}
	retryer.doneChan = make(chan struct{})
	err := retryer.SetRetryStrategy(strategy)
	if err != nil {
		return nil, err
	}
	retryer.retryEventChan = make(chan RetryEvent, strategy.MaxRetrySizeInQueue)
	retryer.retryChan = make(chan *retryEntry, strategy.MaxRetrySizeInQueue)
	retryer.consumeRetryChan()
	retryer.tick()
	return retryer, nil
}

func (d *defaultRetryer) SetRetryStrategy(strategy *RetryStrategy) error {
	if strategy.GrowthRate < 0 {
		return errors.New("growthRate must greater equals 0")
	}
	d.RetryStrategy = strategy
	return nil
}

func (d *defaultRetryer) Invoke(do Do) error {
	err := d.invoke(do)
	if IsRetryError(err) && d.MaxRetryTimes > 0 {
		// first time of timeout
		entry := d.getRetryEntry(do, 0)
		d.timeoutFn(entry)
	}

	return err
}

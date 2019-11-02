package retryer

import (
	"context"
	"errors"
	"fmt"
	recover2 "github.com/blademainer/commons/pkg/recover"
	"time"
)

// Strategies when retry queue is full
type DiscardStrategy int

const (
	DiscardStrategyEarliest  DiscardStrategy = iota + 1 // discard the head of retry queue
	DiscardStrategyLast                                 // discard the last element in retry queue
	DiscardStrategyRejectNew                            // reject new retry job
)

type Do func(ctx context.Context) error

type RetryStrategy struct {
	Timeout             time.Duration
	Interval            time.Duration
	GrowthRate          float32 // rate of growth retry delay.
	MaxRetrySizeInQueue uint16  // Max size of retry function in queue
	DiscardStrategy     DiscardStrategy
}

func NewDefaultDoubleGrowthRateRetryStrategy() *RetryStrategy {
	return NewDoubleGrowthRateRetryStrategy(5*time.Second, 5*time.Second, 1024, DiscardStrategyEarliest)
}

func NewDoubleGrowthRateRetryStrategy(timeout time.Duration, interval time.Duration, maxRetrySizeInQueue uint16, discardStrategy DiscardStrategy) *RetryStrategy {
	strategy := &RetryStrategy{}
	strategy.Timeout = timeout
	strategy.Interval = interval
	strategy.MaxRetrySizeInQueue = maxRetrySizeInQueue
	strategy.DiscardStrategy = discardStrategy
	strategy.GrowthRate = 2.0
	return strategy
}

type Retryer interface {
	SetRetryStrategy(strategy *RetryStrategy) error
	Invoke(do Do) error
}

type defaultRetryer struct {
	*RetryStrategy

	lastDelay      time.Duration
	nextInvokeTime time.Time
	retryChan      chan Do
}

func NewRetryer(strategy *RetryStrategy) {
	retryer := &defaultRetryer{}
	retryer.RetryStrategy = strategy
}

func (d *defaultRetryer) SetRetryStrategy(strategy *RetryStrategy) error {
	if strategy.GrowthRate < 0 {
		return errors.New("growthRate must greater equals 0")
	}
	d.RetryStrategy = strategy
	return nil
}

func (d *defaultRetryer) Invoke(do Do) error {
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
			// timeout
			d.timeoutFn(do)
			err := fmt.Errorf("context deadline exceeded and we should retry later")
			return err
		case <-doneCh:
			// done
			return err
		}
	}
}

func (d *defaultRetryer) timeoutFn(do Do) {
	select {
	case d.retryChan <- do:
	default:
		// chan is fulls
	}
}

package retryer

import "time"

type RetryTimeCalculator interface {
	NextRetryDelayNanoseconds(interval time.Duration, retryTimes int) int64
}

type RetryStrategy struct {
	GrowthRate      float32 // rate of growth retry delay.
	DiscardStrategy DiscardStrategy
}

func NewDefaultDoubleGrowthRateRetryStrategy() *RetryStrategy {
	return NewDefaultRetryStrategy(2.0)
}

func NewDefaultRetryStrategy(growthRate float32) *RetryStrategy {
	strategy := &RetryStrategy{}
	//strategy.Timeout = timeout
	//strategy.Interval = interval
	//strategy.MaxRetrySizeInQueue = maxRetrySizeInQueue
	//strategy.DiscardStrategy = discardStrategy
	//strategy.MaxRetryTimes = maxRetryTimes
	strategy.GrowthRate = growthRate
	return strategy
}

func (s *RetryStrategy) NextRetryDelayNanoseconds(interval time.Duration, retryTimes int) int64 {
	nextRetryNanoSeconds := nextRetryDelayNanoseconds(interval, retryTimes, s.GrowthRate)
	return nextRetryNanoSeconds
}

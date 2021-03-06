package retryer

import "time"

type RetryTimeCalculator interface {
	NextRetryDelayNanoseconds(interval time.Duration, retryTimes int) int64
}

type RetryStrategy struct {
	GrowthRate      int // rate of growth retry delay.
}

func NewDefaultDoubleGrowthRateRetryStrategy() *RetryStrategy {
	return NewDefaultRetryStrategy(2)
}

func NewDefaultRetryStrategy(growthRate int) *RetryStrategy {
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

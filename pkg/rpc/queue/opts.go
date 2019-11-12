package queue

import (
	"github.com/blademainer/commons/pkg/logger"
	"time"
)

type Options struct {
	awaitResponse  bool
	invokeTimeout  time.Duration
	tickInterval   time.Duration
	awaitQueueSize int
}

func NewOptions() *Options {
	options := &Options{}
	options.invokeTimeout = 5 * time.Second
	options.awaitResponse = false
	options.awaitQueueSize = 1024
	options.tickInterval = 1 * time.Second
	return options
}

func (o *Options) AwaitResponse(await bool) *Options {
	o.awaitResponse = await
	return o
}

func (o *Options) InvokeTimeout(invokeTimeout time.Duration) *Options {
	o.invokeTimeout = invokeTimeout
	return o
}

func (o *Options) TickInterval(tickInterval time.Duration) *Options {
	o.tickInterval = tickInterval
	return o
}

func (o *Options) AwaitQueueSize(awaitQueueSize int) *Options {
	if awaitQueueSize < 0 {
		logger.Fatal("awaitQueueSize must greater than 0")
	}
	o.awaitQueueSize = awaitQueueSize
	return o
}

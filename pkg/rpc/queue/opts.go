package queue

import (
	"github.com/blademainer/commons/pkg/logger"
	"time"
)

type Options struct {
	awaitResponse               bool
	invokeTimeout               time.Duration
	tickInterval                time.Duration
	awaitQueueSize              int
	messageHandleConcurrentSize int
}

func NewOptions() *Options {
	options := &Options{}
	options.invokeTimeout = 5 * time.Second
	options.awaitResponse = false
	options.awaitQueueSize = 1024
	options.tickInterval = 1 * time.Second
	options.messageHandleConcurrentSize = 16
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

func (o *Options) MessageHandleConcurrentSize(messageHandleConcurrentSize int) *Options {
	if messageHandleConcurrentSize < 0 {
		logger.Fatal("awaitQueueSize must greater than 0")
	}
	o.messageHandleConcurrentSize = messageHandleConcurrentSize
	return o
}

type InvokeOptions struct {
	produceFunc func(payload []byte) error
}

func NewInvokeOptions() *InvokeOptions {
	options := &InvokeOptions{}
	return options
}

func (options *InvokeOptions) WithProduceFunc(produceFunc func(payload []byte) error) *InvokeOptions {
	options.produceFunc = produceFunc
	return options
}

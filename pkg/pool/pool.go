package pool

import (
	"errors"
	"fmt"
)

type WorkerPool interface {
	Execute(fn func() error)
	Stop() error
}

type defaultWorkerPool struct {
	currency int
	worker   chan func() error
	result   chan error
	doneCh   chan error
}

func InitWorkerPool(concurrent int) (p WorkerPool, doneCh <-chan error, e error) {
	if concurrent < 0 {
		e = errors.New("concurrent must greater than 0")
		return
	}
	pool := &defaultWorkerPool{
		currency: concurrent,
		worker:   make(chan func() error, concurrent),
		result:   make(chan error, concurrent*2),
		doneCh:   make(chan error, concurrent*10),
	}
	doneCh = pool.doneCh
	pool.initWorker()
	p = pool

	return
}

func (p *defaultWorkerPool) initWorker() {
	for i := 0; i < p.currency; i++ {
		go func() {
			for {
				f, ok := <-p.worker
				if !ok {
					fmt.Println("Stopped!!!")
					break
				}
				e := f()
				p.doneCh <- e
				p.result <- e
			}
		}()
	}

	for i := 0; i < p.currency; i++ {
		p.result <- nil
	}
}

func (p *defaultWorkerPool) Execute(fn func() error) {
	e, ok := <-p.result
	if !ok {
		fmt.Println("Stopped!!!")
		return
	}
	if e != nil {
		fmt.Println("Error: ", e.Error())
	}
	p.worker <- fn
}

func (p *defaultWorkerPool) Stop() error {
	close(p.worker)
	close(p.result)
	close(p.doneCh)
	return nil
}

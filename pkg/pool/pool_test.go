package processor

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestDefaultWorkerPool_Execute(t *testing.T) {
	p, d, e := InitWorkerPool(10)
	if e != nil {
		panic(e)
	}

	wg := sync.WaitGroup{}
	wg.Add(1000)
	go func() {
		for i := 0; i < 1000; i++ {
			<-d
			fmt.Println("Done thread.")
			wg.Done()
		}
	}()

	//wg := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		//wg.Add(1)
		j := i
		p.Execute(func() error {
			fmt.Println("Start: ", j)
			time.Sleep(10 * time.Millisecond)
			fmt.Println("Done: ", j)
			return nil
		})
	}

	wg.Wait()

	e = p.Stop()
	if e != nil {
		panic(e)
	}
}

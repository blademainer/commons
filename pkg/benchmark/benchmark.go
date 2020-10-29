package benchmark

import (
	"context"
	"fmt"
	"sync"
)

// Do 执行对象
type Do func(ctx context.Context) error

// Result 执行结果
type Result int

// 执行结果
const (
	ResultOk Result = iota
	ResultError
)

// BenchMark 性能测试
type BenchMark struct {
	Concurrency int
	Count       int
	Do          Do
}

// New 创建性能测试对象
func New(concurrency int, count int, do Do) *BenchMark {
	b := &BenchMark{
		Concurrency: concurrency,
		Count:       count,
		Do:          do,
	}
	return b
}

// Start 开始测试
func (b *BenchMark) Start(ctx context.Context) map[Result]int {
	result := make(map[Result]int)
	resultChan := make(chan Result, b.Concurrency)
	wg := sync.WaitGroup{}
	for i := 0; i < b.Concurrency; i++ {
		wg.Add(1)
		go func() {
			for i := 0; i < b.Count; i++ {
				select {
				case <-ctx.Done():
					fmt.Println("break!!!")
					return
				default:
					err := b.Do(ctx)
					if err != nil {
						resultChan <- ResultError
					} else {
						resultChan <- ResultOk
					}
				}
			}
			wg.Done()
		}()
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("break!!!")
				return
			case r := <-resultChan:
				_, ok := result[r]
				if !ok {
					result[r] = 1
				} else {
					result[r]++
				}
			}
		}
	}()

	wg.Wait()

	return result
}

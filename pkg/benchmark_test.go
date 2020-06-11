package benchmark

import (
	"context"
	"fmt"
	"sync/atomic"
)

func ExampleBenchMark_Start() {
	var i int32 = 0
	b := New(100, 10, func(ctx context.Context) error {
		j := atomic.AddInt32(&i, 1)
		if j%2 == 0 {
			return fmt.Errorf("just error")
		}
		return nil
	})
	rootContext, cancel := context.WithCancel(context.Background())
	defer cancel()
	start := b.Start(rootContext)
	fmt.Printf("%#v\n", start)
	// Output:
	// map[benchmark.Result]int{0:500, 1:500}
}

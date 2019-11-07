package retryer

import (
	"context"
	"fmt"
	"github.com/blademainer/commons/pkg/logger"
	assert "github.com/stretchr/testify/assert"
	"sort"
	"sync/atomic"
	"testing"
	"time"
)

func ExampleNewRetryer() {
	logger.SetLevel(logger.LOG_LEVEL_INFO)
	//os.Setenv(logger.ENV_LOG_LEVEL, logger.LOG_LEVEL_DEBUG)

	//strategy := NewDefaultDoubleGrowthRateRetryStrategy()
	fmt.Println(logger.GetLevel())

	strategy := NewDefaultDoubleGrowthRateRetryStrategy()

	retryer, e := NewRetryer(strategy, 10, 100, 5*time.Millisecond, 5*time.Millisecond, DiscardStrategyEarliest)
	if e != nil {
		panic(e)
	}

	event := retryer.GetEvent()
	go func() {
		for {
			select {
			case e, running := <-event:
				if !running {
					fmt.Println("stopped!")
					return
				}
				fmt.Printf("event: %v running: %v\n", e, running)
			}
		}
	}()

	index := int32(0)
	e = retryer.Invoke(func(ctx context.Context) error {
		fmt.Println("start...", index)
		time.Sleep(10 * time.Millisecond)
		fmt.Println("finish...", index)
		//index++
		atomic.AddInt32(&index, 1)
		return nil
	})
	time.Sleep(1 * time.Second)
	e = retryer.Stop()
	if e != nil {
		fmt.Println(e.Error())
	}
}

func Test_defaultRetryer_Invoke(t *testing.T) {
	logger.SetLevel(logger.LOG_LEVEL_DEBUG)
	//os.Setenv(logger.ENV_LOG_LEVEL, logger.LOG_LEVEL_DEBUG)

	//strategy := NewDefaultDoubleGrowthRateRetryStrategy()
	fmt.Println(logger.GetLevel())

	strategy := NewDefaultDoubleGrowthRateRetryStrategy()

	retryer, e := NewRetryer(strategy, 10, 100, 50*time.Millisecond, 50*time.Millisecond, DiscardStrategyEarliest)
	if e != nil {
		t.Fatal(e)
	}

	event := retryer.GetEvent()
	go func() {
		for {
			select {
			case e, running := <-event:
				if !running {
					fmt.Println("stopped!")
					return
				}
				fmt.Printf("event: %v running: %v\n", e, running)
			}
		}
	}()

	index := int32(0)
	e = retryer.Invoke(func(ctx context.Context) error {
		fmt.Println("start...", index)
		time.Sleep(100 * time.Millisecond)
		fmt.Println("finish...", index)
		//index++
		atomic.AddInt32(&index, 1)
		return nil
	})
	time.Sleep(5 * time.Second)
	retryer.Stop()
	if e != nil {
		fmt.Println(e.Error())
	}
}


func Test_defaultRetryer_insertRetryEntry(t *testing.T) {
	target0, e := time.Parse(time.RFC3339, "2006-01-02T15:04:05+08:00")
	target, e := time.Parse(time.RFC3339, "2006-01-02T15:04:15+08:00")
	target1, e := time.Parse(time.RFC3339, "2006-01-02T15:04:25+08:00")
	target2, e := time.Parse(time.RFC3339, "2006-01-02T15:04:35+08:00")
	if e != nil {
		t.Fatal(e)
		return
	}

	retryer := &defaultRetryer{
		retryEntries: []*retryEntry{
			{nextInvokeTime: target1},
			{nextInvokeTime: target2},
			{nextInvokeTime: target},
			{nextInvokeTime: target0},
		},
	}

	for _, e := range retryer.retryEntries {
		fmt.Print(e.nextInvokeTime.Format(time.RFC3339), ", ")
	}
	fmt.Println()
	sort.Sort(retryer)
	assert.Equal(t, retryer.retryEntries[0].nextInvokeTime, target0)
	assert.Equal(t, retryer.retryEntries[1].nextInvokeTime, target)
	assert.Equal(t, retryer.retryEntries[2].nextInvokeTime, target1)
	assert.Equal(t, retryer.retryEntries[3].nextInvokeTime, target2)
	for _, e := range retryer.retryEntries {
		fmt.Print(e.nextInvokeTime.Format(time.RFC3339), ", ")
	}
	fmt.Println()

}

func Test_defaultRetryer_discard(t *testing.T) {
	entries := make([]*retryEntry, 1024)
	now := time.Now().UnixNano()
	for i := 0; i < 1024; i++ {
		entries[i] = &retryEntry{retryTimes: i, nextInvokeTime: time.Unix(0, now+int64(i))}
	}
	strategy := NewDefaultDoubleGrowthRateRetryStrategy()

	retryer, _ := NewDoubleGrowthRetryer(10 * time.Second)
	r := retryer.(*defaultRetryer)
	r.retryEntries = entries
	fmt.Println("last one: ", r.retryEntries[len(r.retryEntries)-1])

	entry := &retryEntry{retryTimes: 111, nextInvokeTime: time.Unix(0, now+int64(0))}

	strategy.DiscardStrategy = DiscardStrategyRejectNew
	r.discard(entry)
	assert.NotEqual(t, entries[0], entry)
	for _, e := range r.retryEntries {
		fmt.Print(e.retryTimes, ", ")
	}
	fmt.Println()

	strategy.DiscardStrategy = DiscardStrategyLatest
	r.discard(entry)
	assert.Equal(t, r.retryEntries[0], entry)
	for _, e := range r.retryEntries {
		fmt.Print(e.retryTimes, ", ")
	}
	fmt.Println()

	strategy.DiscardStrategy = DiscardStrategyEarliest
	r.discard(entry)
	assert.Equal(t, r.retryEntries[0], entry)
	assert.Equal(t, r.retryEntries[1], entry)
	assert.Equal(t, r.retryEntries[len(r.retryEntries)-1].retryTimes, 1022)
	for _, e := range r.retryEntries {
		fmt.Print(e.retryTimes, ", ")
	}
	fmt.Println()

}

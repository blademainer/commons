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

func Test_defaultRetryer_Invoke(t *testing.T) {
	logger.SetLevel(logger.LOG_LEVEL_INFO)
	//os.Setenv(logger.ENV_LOG_LEVEL, logger.LOG_LEVEL_DEBUG)

	//strategy := NewDefaultDoubleGrowthRateRetryStrategy()
	fmt.Println(logger.GetLevel())

	strategy := NewDoubleGrowthRateRetryStrategy(5*time.Millisecond, 5*time.Millisecond, 10, 100, DiscardStrategyEarliest)

	retryer, e := NewRetryer(strategy)
	if e != nil {
		t.Fatal(e)
	}

	event := retryer.GetEvent()
	go func() {
		for {
			select {
			case e := <-event:
				fmt.Println("event: ", e)
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
	retryer.Stop()
	if e != nil {
		fmt.Println(e.Error())
	}
}

func Test_mul(t *testing.T) {
	type args struct {
		now        time.Time
		interval   time.Duration
		retryTimes int
		growth     float32
	}
	format := time.Now().Format(time.RFC3339)
	fmt.Println(format)
	from, e := time.Parse(time.RFC3339, "2006-01-02T15:04:05+08:00")
	if e != nil {
		t.Fatal(e)
	}
	target, e := time.Parse(time.RFC3339, "2006-01-02T15:04:15+08:00")
	if e != nil {
		t.Fatal(e)
	}
	target1, e := time.Parse(time.RFC3339, "2006-01-02T15:04:25+08:00")
	if e != nil {
		t.Fatal(e)
	}
	target2, e := time.Parse(time.RFC3339, "2006-01-02T15:04:35+08:00")
	if e != nil {
		t.Fatal(e)
	}
	fmt.Println("want...", target2.UnixNano())
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "add0",
			args: args{from, 5 * time.Second, 0, 2.0},
			want: target.UnixNano(),
		},
		{
			name: "add1",
			args: args{from, 5 * time.Second, 1, 2.0},
			want: target1.UnixNano(),
		},
		{
			name: "add2",
			args: args{from, 5 * time.Second, 2, 2.0},
			want: target2.UnixNano(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nextRetryNanoSeconds(tt.args.now, tt.args.interval, tt.args.retryTimes, tt.args.growth); got != tt.want {
				t.Errorf("nextRetryNanoSeconds() = %v(%v), want %v((%v))", got, time.Unix(0, got).Format(time.RFC3339), tt.want, time.Unix(0, tt.want).Format(time.RFC3339))
			}
		})
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

	retryer, _ := NewRetryer(strategy)
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

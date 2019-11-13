package retryer

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"
)

func Test_defaultRetryer_subset(t *testing.T) {
	type fields struct {
		RWMutex        sync.RWMutex
		RetryStrategy  *RetryStrategy
		retryChan      chan *retryEntry
		retryEntries   []*retryEntry
		retryEventChan chan RetryEvent
		doneChan       chan struct{}
	}
	type args struct {
		subIndex int
	}

	entries := make([]*retryEntry, 1024)
	now := time.Now().UnixNano()
	for i := 0; i < 1024; i++ {
		entries[i] = &retryEntry{retryTimes: i, nextInvokeTime: time.Unix(0, now+int64(i))}
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*retryEntry
	}{
		{
			"sub zero",
			fields{
				retryEntries: entries,
			},
			args{subIndex: 0},
			nil,
		},
		{
			"sub 1",
			fields{
				retryEntries: entries,
			},
			args{subIndex: 1},
			entries[:1],
		},
		{
			"sub half",
			fields{
				retryEntries: entries,
			},
			args{subIndex: len(entries) / 2},
			entries[:len(entries)/2],
		},
		{
			"sub all",
			fields{
				retryEntries: entries,
			},
			args{subIndex: len(entries)},
			entries,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &defaultRetryer{
				RWMutex: tt.fields.RWMutex,
				//RetryStrategy:  tt.fields.RetryStrategy,
				retryChan:      tt.fields.retryChan,
				retryEntries:   tt.fields.retryEntries,
				retryEventChan: tt.fields.retryEventChan,
				doneChan:       tt.fields.doneChan,
			}
			if got := d.subset(tt.args.subIndex); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("subset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultRetryer_findBestPos(t *testing.T) {

	target0, e := time.Parse(time.RFC3339, "2006-01-02T15:04:05+08:00")
	target, e := time.Parse(time.RFC3339, "2006-01-02T15:04:15+08:00")
	target1, e := time.Parse(time.RFC3339, "2006-01-02T15:04:25+08:00")
	target2, e := time.Parse(time.RFC3339, "2006-01-02T15:04:35+08:00")

	targetT, e := time.Parse(time.RFC3339, "2006-01-02T15:04:30+08:00")
	targetFirst, e := time.Parse(time.RFC3339, "2006-01-02T15:04:10+08:00")
	targetNone, e := time.Parse(time.RFC3339, "2006-01-02T15:04:00+08:00")
	if e != nil {
		t.Fatal(e)
		return
	}

	retryer := &defaultRetryer{
		retryEntries: []*retryEntry{
			{nextInvokeTime: target0},
			{nextInvokeTime: target},
			{nextInvokeTime: target1},
			{nextInvokeTime: target2},
		},
	}

	for _, e := range retryer.retryEntries {
		fmt.Print(e.nextInvokeTime.Format(time.RFC3339), ", ")
	}

	type args struct {
		now time.Time
	}
	tests := []struct {
		name   string
		fields *defaultRetryer
		args   args
		want   int
	}{
		{
			name:   "first",
			fields: retryer,
			args:   args{targetFirst},
			want:   0,
		},
		{
			name:   "third",
			fields: retryer,
			args:   args{target1},
			want:   2,
		},
		{
			name:   "before",
			fields: retryer,
			args:   args{targetT},
			want:   2,
		},
		{
			name:   "none",
			fields: retryer,
			args:   args{targetNone},
			want:   len(retryer.retryEntries),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fields.findBestPos(tt.args.now); got != tt.want {
				fmt.Printf("tt.name: %v findBestPos() = %v, want %v\n", tt.name, got, tt.want)
				t.Errorf("findBestPos() = %v, want %v", got, tt.want)
			}
		})
	}
}

func find(){

}

func Test_search(t *testing.T) {
	arr := []int{0, 1, 3, 5, 9, 11}
	search := sort.Search(len(arr), func(i int) bool {
		return arr[i] >= 5
	})
	fmt.Println(search)
}

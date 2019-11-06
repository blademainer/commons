package retryer

import (
	"reflect"
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
				RWMutex:        tt.fields.RWMutex,
				RetryStrategy:  tt.fields.RetryStrategy,
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

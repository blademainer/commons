package queue

import (
	"fmt"
	"testing"
	"time"
)

func Test_foundBestPos(t *testing.T) {
	target0, e := time.Parse(time.RFC3339, "2006-01-02T15:04:05+08:00")
	target1, e := time.Parse(time.RFC3339, "2006-01-02T15:04:15+08:00")
	target2, e := time.Parse(time.RFC3339, "2006-01-02T15:04:25+08:00")
	target3, e := time.Parse(time.RFC3339, "2006-01-02T15:04:35+08:00")

	targetT, e := time.Parse(time.RFC3339, "2006-01-02T15:04:30+08:00")
	targetFirst, e := time.Parse(time.RFC3339, "2006-01-02T15:04:10+08:00")
	targetNone, e := time.Parse(time.RFC3339, "2006-01-02T15:04:00+08:00")
	if e != nil {
		t.Fatal(e)
		return
	}

	retryer := &awaitKeeper{
		ttlEntries: []*awaitEntry{
			{ttl: target0},
			{ttl: target1},
			{ttl: target2},
			{ttl: target3},
		},
	}

	for _, e := range retryer.ttlEntries {
		fmt.Print(e.ttl.Format(time.RFC3339), ", ")
	}

	type args struct {
		now time.Time
	}
	tests := []struct {
		name   string
		fields *awaitKeeper
		args   args
		want   int
	}{
		{
			name:   "first",
			fields: retryer,
			args:   args{targetFirst},
			want:   1,
		},
		{
			name:   "third",
			fields: retryer,
			args:   args{target2},
			want:   2,
		},
		{
			name:   "before",
			fields: retryer,
			args:   args{targetT},
			want:   3,
		},
		{
			name:   "none",
			fields: retryer,
			args:   args{targetNone},
			want:   0,
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

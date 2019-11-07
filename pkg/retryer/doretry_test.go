package retryer

import (
	"fmt"
	"testing"
	"time"
)

func Test_nextRetryNanoSeconds(t *testing.T) {
	type args struct {
		now        time.Time
		interval   time.Duration
		retryTimes int
		growth     int
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
	target2, e := time.Parse(time.RFC3339, "2006-01-02T15:04:45+08:00")
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
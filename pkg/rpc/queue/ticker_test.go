package queue

import (
	"fmt"
	"testing"
)

func Test_foundBestPos(t *testing.T) {
	type args struct {
		arr    []int
		search int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "mid",
			args: args{[]int{1, 3, 5, 6}, 4},
			want: 1,
		},
		{
			name: "left",
			args: args{[]int{1, 3, 5, 6}, 1},
			want: 0,
		},
		{
			name: "right",
			args: args{[]int{1, 3, 5, 6}, 6},
			want: 3,
		},
		{
			name: "no",
			args: args{[]int{1, 3, 5, 6}, -1},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := foundBestPos(tt.args.arr, tt.args.search); got != tt.want {
				fmt.Printf("foundBestPos() = %v, want %v \n", got, tt.want)
				t.Errorf("foundBestPos() = %v, want %v", got, tt.want)
			}
		})
	}
}

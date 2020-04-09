package grpc

import (
	"context"
	"fmt"
	"github.com/blademainer/commons/pkg/logger"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"net/url"
	"testing"
)

func ExampleDialUrl() {
	parse, err := url.Parse("https://google.com:443")
	if err != nil {
		logger.Fatal(err)
	}

	dialUrl, err := DialUrl(context.TODO(), *parse)
	if err != nil {
		logger.Fatal(err)
	}
	fmt.Println(dialUrl)
	client := pb.NewGreeterClient(dialUrl)
	r := &pb.HelloRequest{
		Name: "zhangsan",
	}
	bid, err := client.SayHello(context.TODO(), r)
	if err != nil {
		logger.Fatal(err)
	}
	fmt.Println(bid)
}

func TestDialUrl(t *testing.T) {
	parse, err := url.Parse("https://google.com:443")
	if err != nil {
		logger.Fatal(err)
	}

	dialUrl, err := DialUrl(context.TODO(), *parse)
	if err != nil {
		logger.Fatal(err)
	}
	fmt.Println(dialUrl)
}

func TestParseDialTarget(t *testing.T) {
	type args struct {
		target string
	}
	tests := []struct {
		name     string
		args     args
		wantNet  string
		wantAddr string
	}{
		{
			name:     "",
			args:     args{target: "https://google.com:443"},
			wantNet:  "tcp",
			wantAddr: "https://google.com:443",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNet, gotAddr := ParseDialTarget(tt.args.target)
			if gotNet != tt.wantNet {
				t.Errorf("ParseDialTarget() gotNet = %v, want %v", gotNet, tt.wantNet)
			}
			if gotAddr != tt.wantAddr {
				t.Errorf("ParseDialTarget() gotAddr = %v, want %v", gotAddr, tt.wantAddr)
			}
		})
	}
}

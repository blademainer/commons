#!/bin/bash


#cur_dir="`pwd`"
#env GOOS=linux GOARCH=arm go build -o bin/server ./server
# CGO_ENABLED=1
arch="${1:-arm}"
docker run --rm -e GOPROXY=https://goproxy.io -v $GOPATH:/go:rw -w /go/src/github.com/blademainer/commons -e "GO111MODULE=on" -it golang:1.11  bash -c "rm -f go.sum; mkdir -p bin; GOOS=linux GOARCH=${arch}  go build -o bin/${arch}/merger ./pkg/io/merger"

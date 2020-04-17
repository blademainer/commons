#!/bin/bash

export GO111MODULE=on

cmd_dir="demo"

find . -name "*.go" | while read file; do echo "${file%/*}"; done | sort | uniq | while read f; do
  echo "test: $f"
  go test "$f";
done

# Find main() func and build to bin
# For example, build source "cmd/app/main.go" to ./bin/app
grep -Er "func\s+main\(\s*\)" "${cmd_dir}" | awk -F ":" '{print $1}' | while read source; do
  # remove ${cmd_dir} prefix
  dir_name=`echo ${source%/*} | sed "s~${cmd_dir}~~"`
  bin="./bin/$dir_name"
  echo "build source: $source to bin: ${bin}"
  CGO_ENABLED=0 GOOS=linux go build -o ${bin} ./$source
done
#ls -l "$cmd_dir" | egrep '^d' | awk '{print $NF}' | while read source; do
#  CGO_ENABLED=0 GOOS=linux go build -o ./bin/$source ./$cmd_dir/$source
#done

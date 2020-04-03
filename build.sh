#!/usr/bin/env bash
go build -o bin/panic ./demo/panic

find . -name "*.go" | while read file; do echo "${file%/*}"; done | sort | uniq | while read f; do
  echo "test: $f"
  go test "$f";
done

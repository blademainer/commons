#!/usr/bin/env bash
go build -o bin/panic ./demo/panic
go test ./pkg/pool/
go test ./pkg/field/
go test ./pkg/io/
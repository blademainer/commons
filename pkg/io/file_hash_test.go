package io

import (
	"fmt"
	"testing"
)

func TestHashFileMd5(t *testing.T) {
	s, e := HashFileMd5("./file_hash.go")
	if e != nil {
		panic(e)
	}
	fmt.Println(s)
}

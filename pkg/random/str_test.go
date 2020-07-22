package random

import (
	"bytes"
	"fmt"
	"testing"
)

func TestRand_Read(t *testing.T) {
	rand := NewRand()
	bts := make([]byte, 10)
	read, err := rand.Read(bts)
	fmt.Println(read)
	fmt.Println(err)
	fmt.Printf("%s\n", string(bytes.Runes(bts)))
}

func BenchmarkRand_Read(b *testing.B) {
	rand := NewRand()
	bts := make([]byte, 10)
	for i := 0; i < b.N; i++ {
		rand.Read(bts)
	}
}

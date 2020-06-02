package generator

import (
	"fmt"
	"testing"
)

var clusterId = "06"

func TestGenerateOrderId(t *testing.T) {
	generator := New(&clusterId, 1000)
	id := generator.GenerateId()
	fmt.Println(id)
}

var generator = New(&clusterId, 1000)

func BenchmarkGenerateOrderIdAndDuplicate(b *testing.B) {
	result := make(map[string]int)
	for i := 0; i < b.N; i++ {
		id := generator.GenerateId()
		//fmt.Println(id)
		result[id]++
	}
	for k, v := range result {
		if v > 1 {
			fmt.Printf("key: %s value: %d \n", k, v)
		}
	}
}

func BenchmarkGenerateOrderId(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generator.GenerateId()
		//fmt.Println(id)
	}
}

func BenchmarkGenerator_GenerateIndex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generator.GenerateIndex()
	}
}

func BenchmarkSprintf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fmt.Sprintf("%v%v", 1, 2)
	}
}

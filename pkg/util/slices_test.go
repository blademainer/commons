package util

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRemoveElements(t *testing.T) {
	arr := []string{"1", "2", "3", "4"}
	arri := make([]interface{}, len(arr))
	for i, e := range arr {
		arri[i] = e
	}
	elements := RemoveElements(arri, "2", "3")
	fmt.Println(elements)
	assert.Equal(t, "1", elements[0])
	assert.Equal(t, "4", elements[1])
}

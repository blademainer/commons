package collections

import (
	"fmt"
	"testing"
)

func TestSet_Add(t *testing.T) {
	set := NewSet()
	set.Add("1")
	set.Add("2")
	set.Add("3")
	set.Add("3")
	set.Add(1)
	set.Add(nil)
	fmt.Println(set.Entries())
}

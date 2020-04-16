package json

import (
	"fmt"
	"testing"
)

func TestPrettyJson(t *testing.T) {
	type tt struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	t2 := &tt{
		Name: "zhangsan",
		Age:  11,
	}
	json, err := PrettyJson(t2)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(json))
}

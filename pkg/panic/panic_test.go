package panic

import (
	"fmt"
	"testing"
)

func TestPanic(t *testing.T) {
	badFunc()
}

func TestWithRecover(t *testing.T) {
	WithRecover(badFunc2)
}

func TestWithRecoverAndHandle(t *testing.T) {
	WithRecoverAndHandle(badFunc2, func(i interface{}) {
		fmt.Println("Error happened! error: ", i)
	})
}

func badFunc() {
	defer Recover()
	var e error
	fmt.Println("", e.Error())
}

func badFunc2() {
	var e error
	fmt.Println("", e.Error())
}

func ExampleRecover() {
	defer Recover()
	var e error
	fmt.Println("", e.Error())
}

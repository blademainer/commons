package recover

import (
	"fmt"
	"runtime"
	"testing"
)

func TestPanic(t *testing.T) {
	badFunc()
}

func TestWithRecover(t *testing.T) {
	WithRecover(badFunc2)
}

func TestWithRecoverAndHandle(t *testing.T) {
	defer WithRecoverAndHandle(badFunc2, func(i interface{}) {
		fmt.Println("Error happened! error: ", i)
	})

	//panic("just panic")
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

func ExampleWithRecoverAndHandle() {
	defer WithRecoverAndHandle(badFunc2, func(i interface{}) {
		fmt.Println("Error happened! error: ", i)
	})
}

func caller() {
	caller, file, line, ok := runtime.Caller(1)
	name := runtime.FuncForPC(caller).Name()
	fmt.Sprintf("caller: %v file: %v line: %v ok: %v", name, file, line, ok)
	//fmt.Println(s)
}

func TestCaller(t *testing.T) {
	caller()
}

func Benchmark_caller(b *testing.B) {
	for i := 0; i < b.N; i++ {
		caller()
	}
}


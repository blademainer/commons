package main

import "fmt"
import "github.com/blademainer/commons/pkg/panic"

func badFunc() {
	defer panic.Recover()
	var e error
	fmt.Println("", e.Error())
}

func badFunc2() {
	var e error
	fmt.Println("", e.Error())
}

func main() {
	badFunc()
	panic.WithRecover(badFunc2)
	panic.WithRecoverAndHandle(badFunc2, func(i interface{}) {
		fmt.Println("Error happened! error: ", i)
	})
}

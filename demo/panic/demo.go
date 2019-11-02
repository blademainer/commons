package main

import "fmt"
import "github.com/blademainer/commons/pkg/panic"

func badFunc() {
	defer recover.Recover()
	var e error
	fmt.Println("", e.Error())
}

func badFunc2() {
	var e error
	fmt.Println("", e.Error())
}

func main() {
	badFunc()
	recover.WithRecover(badFunc2)
	recover.WithRecoverAndHandle(badFunc2, func(i interface{}) {
		fmt.Println("Error happened! error: ", i)
	})
}

package main

import "fmt"
import "github.com/blademainer/commons/pkg/recoverable"

func badFunc() {
	defer recoverable.Recover()
	var e error
	fmt.Println("", e.Error())
}

func badFunc2() {
	var e error
	fmt.Println("", e.Error())
}

func main() {
	badFunc()
	recoverable.WithRecover(badFunc2)
	recoverable.WithRecoverAndHandle(badFunc2, func(i interface{}) error {
		fmt.Println("Error happened! error: ", i)
		return nil
	})
}

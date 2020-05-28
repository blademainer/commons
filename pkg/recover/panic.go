package recover

import (
	"fmt"
	"runtime/debug"
)

type PanicHandle func(interface{})

//func PrintStack() {
//	var buf [4096]byte
//	n := runtime.Stack(buf[:], false)
//	fmt.Printf("==> %s\n", string(buf[:n]))
//}

func Recover() {
	if err := recover(); err != nil {
		fmt.Println(err) // 这里的err其实就是panic传入的内容，55
		debug.PrintStack()
	}
}

func WithRecover(fn func()) {
	defer Recover()
	fn()
}

func WithRecoverAndHandle(fn func(), handle PanicHandle) {
	defer RecoverWithHandle(handle)
	fn()
}

func RecoverWithHandle(handle PanicHandle) {
	if err := recover(); err != nil {
		handle(err)
	}
}

package recoverable

import (
	"context"
	"fmt"
	"runtime/debug"
)

// RecoveryHandlerFunc is a function that recovers from the panic `p` by returning an `error`.
type RecoveryHandlerFunc func(p interface{}) (err error)

// RecoveryHandlerFuncContext is a function that recovers from the panic `p` by returning an `error`.
// The context can be used to extract request scoped metadata and context values.
type RecoveryHandlerFuncContext func(ctx context.Context, p interface{}) (err error)

// Recover 手动恢复
func Recover() {
	if err := recover(); err != nil {
		fmt.Println(err) // 这里的err其实就是panic传入的内容
		debug.PrintStack()
	}
}

// WithRecover 带自动恢复机制的Recover
func WithRecover(fn func()) {
	defer Recover()
	fn()
}

// WithRecoverAndHandle 带自动恢复机制的Recover，并传入处理错误的处理器
func WithRecoverAndHandle(fn func(), handle RecoveryHandlerFunc) {
	defer RecoverWithHandle(handle)
	fn()
}

// RecoverWithHandle 恢复并传入处理错误的处理器
func RecoverWithHandle(handle RecoveryHandlerFunc) {
	if p := recover(); p != nil {
		err := handle(p)
		fmt.Printf("failed to handle panic: %v error: %v \n", p, err)
	}
}

// WithRecoveryHandlerContext 带自动恢复机制的Recover，并传入处理错误的处理器
func WithRecoveryHandlerContext(ctx context.Context, fn func(), handle RecoveryHandlerFuncContext) {
	defer RecoverWithHandlerContext(ctx, handle)
	fn()
}

// RecoverWithHandlerContext 恢复并传入处理错误的处理器
func RecoverWithHandlerContext(ctx context.Context, handle RecoveryHandlerFuncContext) {
	if p := recover(); p != nil {
		err := handle(ctx, p)
		fmt.Printf("failed to handle panic: %v error: %v \n", p, err)
	}
}

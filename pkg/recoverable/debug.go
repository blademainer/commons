package recoverable

import (
	"fmt"
	"runtime"
)

// Stack 打印当前栈
func Stack() string {
	var buf [4096]byte
	n := runtime.Stack(buf[:], false)
	return fmt.Sprintf("==> %s\n", string(buf[:n]))
}

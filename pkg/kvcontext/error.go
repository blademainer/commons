package kvcontext

// NotKvContextError "非kv类型上下文"的错误类型
type NotKvContextError struct {
	Message string
}

func (n *NotKvContextError) Error() string {
	return n.Message
}

// IsNotKvContextError 判断是否是"非kv类型上下文"的错误类型
func IsNotKvContextError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*NotKvContextError)
	return ok
}

package retryer

type RetryError struct {
	InnerError error
}

func (r *RetryError) Error() string {
	return r.InnerError.Error()
}

func IsRetryError(err error) bool {
	_, cast := err.(*RetryError)
	return cast
}

type LimitedError struct {
	InnerError error
}

func (r *LimitedError) Error() string {
	return r.InnerError.Error()
}

func IsLimitedError(err error) bool {
	_, cast := err.(*LimitedError)
	return cast
}

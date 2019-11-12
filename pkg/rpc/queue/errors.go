package queue

type SkipAwaitError struct {
	message string
}

type TimeoutError struct {
	message string
}

type QueueFulledError struct {
	message string
}

func (t *TimeoutError) Error() string {
	return t.message
}

func (s *SkipAwaitError) Error() string {
	return s.message
}
func (s *QueueFulledError) Error() string {
	return s.message
}

func IsSkipAwaitError(e error) bool {
	_, castOk := e.(*SkipAwaitError)
	return castOk
}

func IsQueueFulledError(e error) bool {
	_, castOk := e.(*QueueFulledError)
	return castOk
}

func IsTimeoutError(e error) bool {
	_, castOk := e.(*TimeoutError)
	return castOk
}

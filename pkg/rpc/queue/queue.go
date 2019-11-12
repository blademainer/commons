package queue

type Queue interface {
	Produce(payload []byte) error
	Consume() (payload []byte, e error)
	Close() (e error)
}

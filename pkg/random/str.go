package random

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

var seq uint64

// Rand rand
type Rand struct {
	Pool *sync.Pool
}

// NewRand init Rand
func NewRand() *Rand {
	p := &sync.Pool{New: func() interface{} {
		return rand.New(rand.NewSource(getSeed()))
	},
	}
	mrand := &Rand{
		Pool: p,
	}
	return mrand
}

// get seed
func getSeed() int64 {
	//1592564536479936000
	seed := (atomic.AddUint64(&seq, 1)%1000 + 1000) * 1e15
	tn := time.Now().UnixNano() % 1e15
	return int64(seed) + tn
}

func (s *Rand) getrand() *rand.Rand {
	return s.Pool.Get().(*rand.Rand)
}
func (s *Rand) putrand(r *rand.Rand) {
	s.Pool.Put(r)
}
func (s *Rand) Read(p []byte) (int, error) {
	r := s.getrand()
	defer s.putrand(r)

	return r.Read(p)
}

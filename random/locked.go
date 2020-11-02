package random

import (
	"math/rand"
	"sync"
)

// NewLocked returns a rand.Rand with a random source protected by a mutex. To be used for concurrent access.
func NewLocked() *rand.Rand {
	return rand.New(NewLockedSource(NewSource())) //nolint:gosec
}

// NewLockedSource wraps the Source with a mutex. To be used for concurrent access.
func NewLockedSource(src rand.Source) rand.Source {
	return &lockedSource{src: src}
}

type lockedSource struct {
	l   sync.Mutex
	src rand.Source
}

func (s *lockedSource) Int63() int64 {
	s.l.Lock()
	defer s.l.Unlock()

	return s.src.Int63()
}

func (s *lockedSource) Seed(seed int64) {
	s.l.Lock()
	defer s.l.Unlock()

	s.src.Seed(seed)
}

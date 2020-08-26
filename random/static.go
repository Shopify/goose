package random

import "math/rand"

// NewStatic returns a rand.Rand that always returns the seeded value. Useful for tests.
func NewStatic(value int64) *rand.Rand {
	return rand.New(NewStaticSource(value))
}

// NewStaticSource returns a rand.Source that always returns the seeded value. Useful for tests.
func NewStaticSource(value int64) rand.Source {
	return &staticSource{value}
}

type staticSource struct {
	value int64
}

func (s *staticSource) Int63() int64 {
	return s.value
}

func (s *staticSource) Seed(seed int64) {
	s.value = seed
}

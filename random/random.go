package random

import (
	"math/rand"
	"time"
)

// New returns a rand.Rand seeded with the current time with nanoseconds precision.
func New() *rand.Rand {
	return rand.New(NewSource()) //nolint:gosec
}

// NewSource returns a rand.Source seeded with the current time with nanoseconds precision.
func NewSource() rand.Source {
	return rand.NewSource(time.Now().UnixNano())
}

package random

import "math/rand"

// NewDummy returns a rand.Rand seeded with 0.
func NewDummy() *rand.Rand {
	return rand.New(NewDummySource())
}

// NewDummySource returns a rand.Source seeded with 0.
func NewDummySource() rand.Source {
	return rand.NewSource(0)
}

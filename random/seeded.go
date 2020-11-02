package random

import "math/rand"

// NewSeeded returns a seeded rand.Rand.
func NewSeeded(seed int64) *rand.Rand {
	return rand.New(NewSeededSource(seed)) //nolint:gosec
}

// NewSeededSource returns a seeded rand.Source, just like the rand package.
// Provided for API completeness
func NewSeededSource(seed int64) rand.Source {
	return rand.NewSource(seed)
}

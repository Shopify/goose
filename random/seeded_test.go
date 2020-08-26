package random

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSeeded(t *testing.T) {
	s := NewSeeded(123)
	// Predictable output
	require.Equal(t, uint64(0x4a68998bed5c40f1), s.Uint64())
	require.Equal(t, uint64(0x835b51599210f9ba), s.Uint64())
}

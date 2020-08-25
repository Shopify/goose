package random

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewDummy(t *testing.T) {
	s := NewDummy()
	// Predictable output
	require.Equal(t, uint64(0x78fc2ffac2fd9401), s.Uint64())
	require.Equal(t, uint64(0x1f5b0412ffd341c0), s.Uint64())
}

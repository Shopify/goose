package random

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewLockedSource(t *testing.T) {
	s := NewMockSource()
	l := NewLockedSource(s)

	s.On("Int63").Return(int64(123)).Once()

	require.Equal(t, int64(123), l.Int63())

	s.AssertExpectations(t)
}

func TestNewLocked(t *testing.T) {
	l := NewLocked()
	require.NotEqual(t, l.Uint64(), l.Uint64())
}

package random

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewStatic(t *testing.T) {
	s := NewStatic(123)
	require.Equal(t, int64(123), s.Int63())
}

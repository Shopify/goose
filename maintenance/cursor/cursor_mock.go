package cursor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type MockCursor struct {
	current int
}

func NewMockCursor(current int) *MockCursor {
	return &MockCursor{current}
}

func (c *MockCursor) Current(context.Context) (int, error) {
	return c.current, nil
}

func (c *MockCursor) Increment(context.Context) (int, error) {
	c.current++
	return c.current, nil
}

func (c *MockCursor) Set(ctx context.Context, value int) error {
	c.current = value
	return nil
}

func (c *MockCursor) Assert(t *testing.T, expected int) {
	require.Equal(t, expected, c.current)
}

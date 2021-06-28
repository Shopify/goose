package cursor

import (
	"context"
	"testing"

	"github.com/Shopify/go-cache/v2"
	"github.com/stretchr/testify/require"

	"github.com/Shopify/courier/pkg/testutils"
)

func TestCursor(t *testing.T) {
	ctx := context.Background()

	cursorName := testutils.UniqueString("testcursor")
	c := NewCursor(cursorName, cache.NewMemoryClient())

	index, err := c.Current(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, index)

	index, err = c.Current(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, index)

	index, err = c.Increment(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, index)

	index, err = c.Increment(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, index)

	index, err = c.Current(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, index)

	err = c.Set(ctx, 33)
	require.NoError(t, err)

	index, err = c.Current(ctx)
	require.NoError(t, err)
	require.Equal(t, 33, index)
}

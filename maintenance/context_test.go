package maintenance

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type testCtxKey struct{}

func Test_BackgroundContextWithValues_Propagate_Values(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, testCtxKey{}, "VALUE")

	c := BackgroundContextWithValues(ctx)
	require.Equal(t, "VALUE", c.Value(testCtxKey{}))
}

func Test_BackgroundContextWithValues_Block_Cancellation(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, testCtxKey{}, "VALUE")

	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel()
	require.Equal(t, context.Canceled, cancelledCtx.Err())

	c := BackgroundContextWithValues(cancelledCtx)
	require.NoError(t, c.Err())
}

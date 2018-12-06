package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithUUID(t *testing.T) {
	ctx := context.Background()
	ctx, id := WithUUID(ctx)

	assert.NotEmpty(t, id)
	assert.Equal(t, id, GetLoggableValue(ctx, UUIDKey))

	// Calling a second time has now effect
	ctx2, id2 := WithUUID(ctx)
	assert.Equal(t, ctx, ctx2)
	assert.Equal(t, id, id2)
}

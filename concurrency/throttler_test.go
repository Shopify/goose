package concurrency

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Shopify/goose/timetracker"
)

func TestThrottler(t *testing.T) {
	ctx := context.Background()
	mockLimiter := NewMockLimiter(true)
	mockTracker := timetracker.NewMockTracker(true)

	th := NewThrottler(mockLimiter, mockTracker, 2*time.Second)

	t.Run("free", func(t *testing.T) {
		mockTracker.On("Average").Return(1 * time.Second).Once()
		mockLimiter.On("MaxConcurrency").Return(uint(1)).Once()
		mockLimiter.On("Waiting").Return(int32(2)).Once()

		var ran bool
		err := th.Run(ctx, func() error {
			ran = true
			return nil
		})
		require.NoError(t, err)
		require.True(t, ran)

		mockLimiter.AssertExpectations(t)
		mockTracker.AssertExpectations(t)
	})

	t.Run("busy", func(t *testing.T) {
		mockLimiter.On("MaxConcurrency").Return(uint(1)).Once()
		mockLimiter.On("Waiting").Return(int32(3)).Once()
		mockTracker.On("Average").Return(1 * time.Second).Once()

		err := th.Run(ctx, nil)
		require.Equal(t, &ErrThrottled{3 * time.Second}, err)

		mockLimiter.AssertExpectations(t)
		mockTracker.AssertExpectations(t)
	})
}

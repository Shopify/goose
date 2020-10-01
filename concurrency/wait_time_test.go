package concurrency

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEstimatedWaitTime(t *testing.T) {
	tests := []struct {
		concurrency uint
		waiting     int32
		average     time.Duration
		want        time.Duration
	}{
		{concurrency: 0, waiting: 10, average: 10 * time.Second, want: 0},
		{concurrency: 10, waiting: 0, average: 10 * time.Second, want: 0},
		{concurrency: 10, waiting: 10, average: 0, want: 0},
		{concurrency: 5, waiting: 10, average: 10 * time.Second, want: 20 * time.Second},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v/%v/%v", tt.concurrency, tt.waiting, tt.average), func(t *testing.T) {
			limiter := NewMockLimiter(false)
			limiter.On("MaxConcurrency").Return(tt.concurrency).Maybe()
			limiter.On("Waiting").Return(tt.waiting).Maybe()

			estimate := EstimatedWaitTime(limiter, tt.average)
			require.Equal(t, tt.want, estimate)

			limiter.AssertExpectations(t)
		})
	}
}

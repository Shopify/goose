package concurrency

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// concurrency_0-8         	 1000000	      1891 ns/op
// concurrency_1-8         	  432932	      3447 ns/op
// concurrency_10-8        	  650433	      5317 ns/op
// concurrency_100-8       	  819081	      2538 ns/op
func BenchmarkLimiter(b *testing.B) {
	concurrencies := []uint{NoLimit, 1, 10, 100}

	for _, concurrency := range concurrencies {
		b.Run(fmt.Sprintf("concurrency %d", concurrency), func(b *testing.B) {
			limiter := NewLimiter(concurrency)
			ctx := context.Background()

			wg := sync.WaitGroup{}
			wg.Add(b.N)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				go func() {
					_ = limiter.Run(ctx, func() error {
						time.Sleep(time.Nanosecond) // Force the scheduler for a more meaningful test
						return nil
					})
					wg.Done()
				}()
			}

			wg.Wait()
		})
	}
}

func TestConcurrencyLimiter_cancel(t *testing.T) {
	limiter := NewLimiter(1)
	ctx := context.Background()

	// Queue is empty, it doesn't matter if the context is canceled.
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	require.NoError(t, limiter.Run(ctx, func() error { return nil }))

	started := make(chan struct{})
	wait := make(chan struct{})
	done := make(chan error)
	go func() {
		done <- limiter.Run(ctx, func() error {
			close(started)
			<-wait
			return nil
		})
	}()

	<-started

	err := limiter.Run(cancelCtx, func() error { return nil })
	require.EqualError(t, err, "context canceled")

	close(wait)
	require.NoError(t, <-done)
}

func expectCount(t *testing.T, call func() int32, expected int32) {
	for i := 0; i < 20; i++ {
		if call() == expected {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.Failf(t, "wrong count", "expect count to be %d, but it's stuck at %d", expected, call())
}

func TestConcurrencyLimiter_stats(t *testing.T) {
	limiter := NewLimiter(2)
	ctx := context.Background()

	wait := make(chan struct{})
	done := make(chan error)

	queue := func() {
		done <- limiter.Run(ctx, func() error {
			<-wait
			return nil
		})
	}

	require.Equal(t, int32(0), limiter.Waiting())
	require.Equal(t, int32(0), limiter.Running())

	go queue()
	go queue()
	go queue()

	expectCount(t, limiter.Running, 2)
	expectCount(t, limiter.Waiting, 1)

	wait <- struct{}{}
	require.NoError(t, <-done)
	expectCount(t, limiter.Waiting, 0)
	expectCount(t, limiter.Running, 2)

	wait <- struct{}{}
	require.NoError(t, <-done)
	expectCount(t, limiter.Waiting, 0)
	expectCount(t, limiter.Running, 1)

	wait <- struct{}{}
	require.NoError(t, <-done)

	require.Equal(t, int32(0), limiter.Waiting())
	require.Equal(t, int32(0), limiter.Running())
}

func TestConcurrencyLimiter_random(t *testing.T) {
	runs := 1000
	concurrency := uint(10)

	limiter := NewLimiter(concurrency)
	require.Equal(t, concurrency, limiter.MaxConcurrency())
	ctx := context.Background()

	done := make(chan error)

	for i := 0; i <= runs; i++ {
		go func() {
			done <- limiter.Run(ctx, func() error {
				time.Sleep(time.Duration(rand.Float64()) * 100 * time.Millisecond)
				return nil
			})
		}()
	}

	for i := 0; i <= runs; i++ {
		require.NoError(t, <-done)
	}

	require.Equal(t, int32(0), limiter.Waiting())
	require.Equal(t, int32(0), limiter.Running())
}

func TestConcurrencyLimiter_unlimited(t *testing.T) {
	limiter := NewLimiter(0)
	require.Equal(t, uint(0), limiter.MaxConcurrency())
	ctx := context.Background()

	wait := make(chan struct{})
	done := make(chan error)

	queue := func() {
		done <- limiter.Run(ctx, func() error {
			<-wait
			return nil
		})
	}

	go queue()
	go queue()

	expectCount(t, limiter.Waiting, 0)
	expectCount(t, limiter.Running, 2)

	wait <- struct{}{}
	wait <- struct{}{}

	require.NoError(t, <-done)
	require.NoError(t, <-done)

	require.Equal(t, int32(0), limiter.Waiting())
	require.Equal(t, int32(0), limiter.Running())
}

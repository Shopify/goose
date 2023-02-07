package concurrency

import (
	"context"
	"fmt"
	"time"

	"github.com/Shopify/goose/v2/timetracker"
)

func NewThrottler(limiter Limiter, tracker timetracker.Tracker, waitTimeout time.Duration) Throttler {
	return &throttler{
		limiter:     limiter,
		tracker:     tracker,
		waitTimeout: waitTimeout,
	}
}

type Throttler interface {
	Run(ctx context.Context, fn func() error) error
}

type ErrThrottled struct {
	WaitTime time.Duration
}

func (t *ErrThrottled) Error() string {
	return fmt.Sprintf("throttled, retry after %.02f seconds", t.WaitTime.Seconds())
}

type throttler struct {
	limiter     Limiter
	tracker     timetracker.Tracker
	waitTimeout time.Duration
}

func (t *throttler) Run(ctx context.Context, fn func() error) error {
	if waitTime := EstimatedWaitTime(t.limiter, t.tracker.Average()); waitTime > t.waitTimeout {
		return &ErrThrottled{waitTime}
	}

	return t.limiter.Run(ctx, func() error {
		defer t.tracker.Start().Finish()

		return fn()
	})
}

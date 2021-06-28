package maintenance

import (
	"context"
	"time"
)

type RateLimiter interface {
	Wait(context.Context) error
}

type SimpleRateLimiter struct {
	interval       time.Duration
	timeOfLastCall time.Time
}

func NewRateLimiter(calls int, interval time.Duration) *SimpleRateLimiter {
	return &SimpleRateLimiter{
		interval: interval / time.Duration(calls),
	}
}

func (l *SimpleRateLimiter) Wait(ctx context.Context) error {
	defer func() { l.timeOfLastCall = time.Now() }()

	if l.timeOfLastCall.IsZero() {
		return nil
	}

	waitTime := l.interval - time.Since(l.timeOfLastCall)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(waitTime):
		return nil
	}
}

type NoLimitRateLimiter struct{}

func (l *NoLimitRateLimiter) Wait(ctx context.Context) error {
	return nil
}

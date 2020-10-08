package concurrency

import (
	"context"
	"sync/atomic"

	"golang.org/x/sync/semaphore"

	"github.com/Shopify/goose/statsd"
)

const (
	NoLimit      uint  = 0
	AlwaysRecord int32 = 0
)

type Limiter interface {
	// Run executes a function, making sure at most a certain number of calls are executing simultaneously
	// It does _not_ start a goroutine, the function will be executing on the caller's stack.
	// If the context is canceled while waiting, it will abort.
	Run(ctx context.Context, fn func() error) error

	// Waiting returns a (racy) counter of how many Run calls are currently waiting to run.
	Waiting() int32

	// Running returns a (racy) counter of how many Run calls are currently executing their function.
	Running() int32

	MaxConcurrency() uint
}

type gaugor interface {
	Gauge(ctx context.Context, n float64, ts ...statsd.Tags)
}

type limiter struct {
	semaphore     *semaphore.Weighted
	concurrency   uint
	waiting       int32
	running       int32
	gauge         gaugor
	tags          statsd.Tags
	sampling      int32
	sampleCounter int32
}

func NewLimiter(concurrency uint) Limiter {
	return NewLimiterWithGauge(concurrency, nil, nil)
}

func NewLimiterWithGauge(concurrency uint, gauge gaugor, tags statsd.Tags) Limiter {
	return NewLimiterWithSampledGauge(concurrency, gauge, 0, tags)
}

func NewGauge(gauge gaugor, tags statsd.Tags) Limiter {
	return NewSampledGauge(gauge, AlwaysRecord, tags)
}

func NewSampledGauge(gauge gaugor, sampling int32, tags statsd.Tags) Limiter {
	return NewLimiterWithSampledGauge(NoLimit, gauge, sampling, tags)
}

// NewLimiterWithSampledGauge creates a Limiter that will publish the a gauge every <sampling> Run
func NewLimiterWithSampledGauge(concurrency uint, gauge gaugor, sampling int32, tags statsd.Tags) Limiter {
	limiter := &limiter{
		concurrency: concurrency,
		gauge:       gauge,
		sampling:    sampling,
		tags:        tags,
	}
	if concurrency > 0 {
		limiter.semaphore = semaphore.NewWeighted(int64(concurrency))
	}
	return limiter
}

func (c *limiter) Run(ctx context.Context, fn func() error) error {
	if c.semaphore != nil {
		if err := c.acquire(ctx); err != nil {
			return err
		}
		defer c.semaphore.Release(1)
	}

	return c.run(ctx, fn)
}

func (c *limiter) acquire(ctx context.Context) error {
	c.deltaAndMaybePublish(ctx, &c.waiting, 1, "waiting")
	defer c.deltaAndMaybePublish(ctx, &c.waiting, -1, "waiting")

	return c.semaphore.Acquire(ctx, 1)
}

func (c *limiter) run(ctx context.Context, fn func() error) error {
	c.deltaAndMaybePublish(ctx, &c.running, 1, "running")
	defer c.deltaAndMaybePublish(ctx, &c.running, -1, "running")

	return fn()
}

func (c *limiter) deltaAndMaybePublish(ctx context.Context, ptr *int32, delta int32, state string) {
	current := atomic.AddInt32(ptr, delta)

	if c.gauge == nil {
		return
	}
	if c.sampling != AlwaysRecord {
		counter := atomic.AddInt32(&c.sampleCounter, 1) // will overflow and loop gracefully.
		if counter%c.sampling > 0 {
			return
		}
	}
	c.gauge.Gauge(ctx, float64(current), statsd.Tags{"state": state}, c.tags)
}

func (c *limiter) Waiting() int32 {
	return atomic.LoadInt32(&c.waiting)
}

func (c *limiter) Running() int32 {
	return atomic.LoadInt32(&c.running)
}

func (c *limiter) MaxConcurrency() uint {
	return c.concurrency
}

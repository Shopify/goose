package statsd

import (
	"context"
	"time"
)

type ForwardHandler func(ctx context.Context, mType string, name string, value interface{}, tags []string, rate float64) error

// NewForwardingBackend creates a new Backend that sends all metrics to a ForwardHandler
func NewForwardingBackend(handler ForwardHandler) Backend {
	return &forwardingBackend{
		handler: handler,
	}
}

type forwardingBackend struct {
	handler ForwardHandler
}

func (b *forwardingBackend) Gauge(ctx context.Context, name string, value float64, tags []string, rate float64) error {
	return b.handler(ctx, "gauge", name, value, tags, rate)
}

func (b *forwardingBackend) Count(ctx context.Context, name string, value int64, tags []string, rate float64) error {
	return b.handler(ctx, "count", name, value, tags, rate)
}

func (b *forwardingBackend) Histogram(ctx context.Context, name string, value float64, tags []string, rate float64) error {
	return b.handler(ctx, "histogram", name, value, tags, rate)
}

func (b *forwardingBackend) Distribution(ctx context.Context, name string, value float64, tags []string, rate float64) error {
	return b.handler(ctx, "distribution", name, value, tags, rate)
}

func (b *forwardingBackend) Set(ctx context.Context, name string, value string, tags []string, rate float64) error {
	return b.handler(ctx, "set", name, value, tags, rate)
}

func (b *forwardingBackend) Timing(ctx context.Context, name string, value time.Duration, tags []string, rate float64) error {
	return b.handler(ctx, "timing", name, value, tags, rate)
}

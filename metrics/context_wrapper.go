package metrics

import (
	"context"
	"time"
)

// NewContextWrapper creates a new Backend that prefixes all metrics
func NewContextWrapper(backend Backend) Backend {
	return &contextWrapper{
		backend: backend,
	}
}

type contextWrapper struct {
	backend Backend
}

func (w *contextWrapper) tags(ctx context.Context, tags Tags) Tags {
	return TagsFromContext(ctx).Merge(tags)
}

func (w *contextWrapper) Gauge(ctx context.Context, name string, value float64, tags Tags, rate float64) error {
	return w.backend.Gauge(ctx, name, value, w.tags(ctx, tags), rate)
}

func (w *contextWrapper) Count(ctx context.Context, name string, value int64, tags Tags, rate float64) error {
	return w.backend.Count(ctx, name, value, w.tags(ctx, tags), rate)
}

func (w *contextWrapper) Histogram(ctx context.Context, name string, value float64, tags Tags, rate float64) error {
	return w.backend.Histogram(ctx, name, value, w.tags(ctx, tags), rate)
}

func (w *contextWrapper) Distribution(ctx context.Context, name string, value float64, tags Tags, rate float64) error {
	return w.backend.Distribution(ctx, name, value, w.tags(ctx, tags), rate)
}

func (w *contextWrapper) Set(ctx context.Context, name string, value string, tags Tags, rate float64) error {
	return w.backend.Set(ctx, name, value, w.tags(ctx, tags), rate)
}

func (w *contextWrapper) Timing(ctx context.Context, name string, value time.Duration, tags Tags, rate float64) error {
	return w.backend.Timing(ctx, name, value, w.tags(ctx, tags), rate)
}

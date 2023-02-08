package metrics

import (
	"context"
	"time"
)

// NewTagsWrapper creates a new Backend that always injects some tags
func NewTagsWrapper(backend Backend, tags Tags) Backend {
	if len(tags) == 0 {
		return backend
	}
	return &tagsWrapper{
		backend: backend,
		tags:    tags,
	}
}

type tagsWrapper struct {
	backend Backend
	tags    Tags
}

func (w *tagsWrapper) Gauge(ctx context.Context, name string, value float64, tags Tags, rate float64) error {
	return w.backend.Gauge(ctx, name, value, w.tags.Merge(tags), rate)
}

func (w *tagsWrapper) Count(ctx context.Context, name string, value int64, tags Tags, rate float64) error {
	return w.backend.Count(ctx, name, value, w.tags.Merge(tags), rate)
}

func (w *tagsWrapper) Histogram(ctx context.Context, name string, value float64, tags Tags, rate float64) error {
	return w.backend.Histogram(ctx, name, value, w.tags.Merge(tags), rate)
}

func (w *tagsWrapper) Distribution(ctx context.Context, name string, value float64, tags Tags, rate float64) error {
	return w.backend.Distribution(ctx, name, value, w.tags.Merge(tags), rate)
}

func (w *tagsWrapper) Set(ctx context.Context, name string, value string, tags Tags, rate float64) error {
	return w.backend.Set(ctx, name, value, w.tags.Merge(tags), rate)
}

func (w *tagsWrapper) Timing(ctx context.Context, name string, value time.Duration, tags Tags, rate float64) error {
	return w.backend.Timing(ctx, name, value, w.tags.Merge(tags), rate)
}

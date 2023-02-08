package metrics

import (
	"context"
	"strings"
	"time"
)

// NewPrefixWrapper creates a new Backend that prefixes all metrics
func NewPrefixWrapper(backend Backend, prefix string) Backend {
	if prefix == "" {
		return backend
	}
	if !strings.HasSuffix(prefix, ".") {
		prefix += "."
	}
	return &prefixWrapper{
		backend: backend,
		prefix:  prefix,
	}
}

type prefixWrapper struct {
	backend Backend
	prefix  string
}

func (w *prefixWrapper) name(name string) string {
	return w.prefix + name
}

func (w *prefixWrapper) Gauge(ctx context.Context, name string, value float64, tags Tags, rate float64) error {
	return w.backend.Gauge(ctx, w.name(name), value, tags, rate)
}

func (w *prefixWrapper) Count(ctx context.Context, name string, value int64, tags Tags, rate float64) error {
	return w.backend.Count(ctx, w.name(name), value, tags, rate)
}

func (w *prefixWrapper) Histogram(ctx context.Context, name string, value float64, tags Tags, rate float64) error {
	return w.backend.Histogram(ctx, w.name(name), value, tags, rate)
}

func (w *prefixWrapper) Distribution(ctx context.Context, name string, value float64, tags Tags, rate float64) error {
	return w.backend.Distribution(ctx, w.name(name), value, tags, rate)
}

func (w *prefixWrapper) Set(ctx context.Context, name string, value string, tags Tags, rate float64) error {
	return w.backend.Set(ctx, w.name(name), value, tags, rate)
}

func (w *prefixWrapper) Timing(ctx context.Context, name string, value time.Duration, tags Tags, rate float64) error {
	return w.backend.Timing(ctx, w.name(name), value, tags, rate)
}

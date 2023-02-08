package metrics

import (
	"context"
	"time"
)

type Metric struct {
	Type  string
	Name  string
	Value interface{}
	Tags  Tags
	Rate  float64
}

type ForwardHandler func(ctx context.Context, metric *Metric) error

// NewForwardingBackend creates a new Backend that sends all metrics to a ForwardHandler
func NewForwardingBackend(handler ForwardHandler) Backend {
	return &forwardingBackend{
		handler: handler,
	}
}

type forwardingBackend struct {
	handler ForwardHandler
}

func (b *forwardingBackend) call(ctx context.Context, method string, name string, value interface{}, tags Tags, rate float64) error {
	return b.handler(ctx, &Metric{
		Type:  method,
		Name:  name,
		Value: value,
		Tags:  tags,
		Rate:  rate,
	})
}

func (b *forwardingBackend) Gauge(ctx context.Context, name string, value float64, tags Tags, rate float64) error {
	return b.call(ctx, "Gauge", name, value, tags, rate)
}

func (b *forwardingBackend) Count(ctx context.Context, name string, value int64, tags Tags, rate float64) error {
	return b.call(ctx, "Count", name, value, tags, rate)
}

func (b *forwardingBackend) Histogram(ctx context.Context, name string, value float64, tags Tags, rate float64) error {
	return b.call(ctx, "Histogram", name, value, tags, rate)
}

func (b *forwardingBackend) Distribution(ctx context.Context, name string, value float64, tags Tags, rate float64) error {
	return b.call(ctx, "Distribution", name, value, tags, rate)
}

func (b *forwardingBackend) Set(ctx context.Context, name string, value string, tags Tags, rate float64) error {
	return b.call(ctx, "Set", name, value, tags, rate)
}

func (b *forwardingBackend) Timing(ctx context.Context, name string, value time.Duration, tags Tags, rate float64) error {
	return b.call(ctx, "Timing", name, value, tags, rate)
}

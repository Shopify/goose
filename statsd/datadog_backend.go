package statsd

import (
	"context"
	"time"

	"github.com/DataDog/datadog-go/statsd"
)

// NewDatadogBackend instantiates a new datadog/statsd connection. When running
// in containers at shopify, the endpoint should generally be "localhost:8125".
//
// `namespace` is an optional prefix to be prepended to every metric submitted.
// It should end with a period to separate it from the metric name.
//
// `tags` is a set of tags that will be included with every metric submitted.
func NewDatadogBackend(endpoint, namespace string, tags []string) (Backend, error) {
	client, err := statsd.New(endpoint)
	if err != nil {
		return nil, err
	}
	client.Namespace = namespace
	client.Tags = tags
	return &datadogBackend{
		client: client,
	}, nil
}

type datadogBackend struct {
	client *statsd.Client
}

func (b *datadogBackend) Gauge(ctx context.Context, name string, value float64, tags []string, rate float64) error {
	return b.client.Gauge(name, value, tags, rate)
}

func (b *datadogBackend) Count(ctx context.Context, name string, value int64, tags []string, rate float64) error {
	return b.client.Count(name, value, tags, rate)
}

func (b *datadogBackend) Histogram(ctx context.Context, name string, value float64, tags []string, rate float64) error {
	return b.client.Histogram(name, value, tags, rate)
}

func (b *datadogBackend) Distribution(ctx context.Context, name string, value float64, tags []string, rate float64) error {
	return b.client.Distribution(name, value, tags, rate)
}

func (b *datadogBackend) Decr(ctx context.Context, name string, tags []string, rate float64) error {
	return b.client.Decr(name, tags, rate)
}

func (b *datadogBackend) Incr(ctx context.Context, name string, tags []string, rate float64) error {
	return b.client.Incr(name, tags, rate)
}

func (b *datadogBackend) Set(ctx context.Context, name string, value string, tags []string, rate float64) error {
	return b.client.Set(name, value, tags, rate)
}

func (b *datadogBackend) Timing(ctx context.Context, name string, value time.Duration, tags []string, rate float64) error {
	return b.client.Timing(name, value, tags, rate)
}

package statsd

import (
	"context"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
)

// NewDatadogBackend instantiates a new datadog/statsd connection. When running
// in containers at shopify, the endpoint should generally be "localhost:8125".
//
// `namespace` is an optional prefix to be prepended to every metric submitted.
// It should end with a period to separate it from the metric name.
//
// `tags` is a set of tags that will be included with every metric submitted.
// STATSD_DEFAULT_TAGS env variable will be read automatically and added to default tags.
func NewDatadogBackend(endpoint, namespace string, tags []string) (Backend, error) {
	defaultTags := append(defaultTagsFromEnv(), tags...)
	client, err := statsd.New(
		endpoint,
		statsd.WithNamespace(namespace),
		statsd.WithTags(defaultTags),
		statsd.WithoutTelemetry(),
		statsd.WithoutOriginDetection(),
		statsd.WithClientSideAggregation(),
	)
	if err != nil {
		return nil, err
	}
	return &datadogBackend{
		client: client,
	}, nil
}

type datadogBackend struct {
	client *statsd.Client
}

func (b *datadogBackend) Gauge(_ context.Context, name string, value float64, tags []string, rate float64) error {
	return b.client.Gauge(name, value, tags, rate)
}

func (b *datadogBackend) Count(_ context.Context, name string, value int64, tags []string, rate float64) error {
	return b.client.Count(name, value, tags, rate)
}

func (b *datadogBackend) Histogram(_ context.Context, name string, value float64, tags []string, rate float64) error {
	return b.client.Histogram(name, value, tags, rate)
}

func (b *datadogBackend) Distribution(_ context.Context, name string, value float64, tags []string, rate float64) error {
	return b.client.Distribution(name, value, tags, rate)
}

func (b *datadogBackend) Set(_ context.Context, name string, value string, tags []string, rate float64) error {
	return b.client.Set(name, value, tags, rate)
}

func (b *datadogBackend) Timing(_ context.Context, name string, value time.Duration, tags []string, rate float64) error {
	return b.client.Timing(name, value, tags, rate)
}

func (b *datadogBackend) Close() error {
	return b.client.Close()
}

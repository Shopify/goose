package metrics

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
// STATSD_DEFAULT_TAGS env variable will be read automatically and added to default tags.
func NewDatadogBackend(endpoint, namespace string, tags Tags) (Backend, error) {
	client, err := statsd.New(endpoint)
	if err != nil {
		return nil, err
	}
	client.Namespace = namespace
	client.Tags = defaultTagsFromEnv().Merge(tags).StringSlice()
	return &datadogBackend{
		client: client,
	}, nil
}

type datadogBackend struct {
	client *statsd.Client
}

func (b *datadogBackend) Gauge(ctx context.Context, name string, value float64, tags Tags, rate float64) error {
	return b.client.Gauge(name, value, tags.StringSlice(), rate)
}

func (b *datadogBackend) Count(ctx context.Context, name string, value int64, tags Tags, rate float64) error {
	return b.client.Count(name, value, tags.StringSlice(), rate)
}

func (b *datadogBackend) Histogram(ctx context.Context, name string, value float64, tags Tags, rate float64) error {
	return b.client.Histogram(name, value, tags.StringSlice(), rate)
}

func (b *datadogBackend) Distribution(ctx context.Context, name string, value float64, tags Tags, rate float64) error {
	return b.client.Distribution(name, value, tags.StringSlice(), rate)
}

func (b *datadogBackend) Set(ctx context.Context, name string, value string, tags Tags, rate float64) error {
	return b.client.Set(name, value, tags.StringSlice(), rate)
}

func (b *datadogBackend) Timing(ctx context.Context, name string, value time.Duration, tags Tags, rate float64) error {
	return b.client.Timing(name, value, tags.StringSlice(), rate)
}

package metrics

import (
	"context"
	"time"
)

// Backend is an interface to a backend instance, typically wrapping a Statsd backend.
//
// The statsd protocol supports more types than we do: we can add these as we
// need them.
type Backend interface {
	// Gauge measures the value of a metric at a particular time.
	Gauge(ctx context.Context, name string, value float64, tags Tags, rate float64) error
	// Count tracks how many times something happened per second.
	Count(ctx context.Context, name string, value int64, tags Tags, rate float64) error
	// Histogram tracks the statistical distribution of a set of values on each host.
	Histogram(ctx context.Context, name string, value float64, tags Tags, rate float64) error
	// Distribution tracks the statistical distribution of a set of values across your infrastructure.
	Distribution(ctx context.Context, name string, value float64, tags Tags, rate float64) error
	// Set counts the number of unique elements in a group.
	Set(ctx context.Context, name string, value string, tags Tags, rate float64) error
	// Timing sends timing information, it is an alias for TimeInMilliseconds
	Timing(ctx context.Context, name string, value time.Duration, tags Tags, rate float64) error
}

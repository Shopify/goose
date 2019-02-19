// Package statsd contains a singleton statsd client for use in all other
// packages. It can be configured once at application startup, and imported by
// any package that wishes to record metrics.
package statsd

import (
	"context"
	"time"

	"github.com/Shopify/goose/logger"
)

var log = logger.New("statsd")

// Backend is an interface to a Statsd instance, currently implemented by
// nullBackend and NewDatadogBackend (go-dogstatsd).
//
// The statsd protocol supports more types than we do: we can add these as we
// need them.
type Backend interface {
	// Gauge measures the value of a metric at a particular time.
	Gauge(ctx context.Context, name string, value float64, tags []string, rate float64) error
	// Count tracks how many times something happened per second.
	Count(ctx context.Context, name string, value int64, tags []string, rate float64) error
	// Histogram tracks the statistical distribution of a set of values on each host.
	Histogram(ctx context.Context, name string, value float64, tags []string, rate float64) error
	// Distribution tracks the statistical distribution of a set of values across your infrastructure.
	Distribution(ctx context.Context, name string, value float64, tags []string, rate float64) error
	// Decr is just Count of -1
	Decr(ctx context.Context, name string, tags []string, rate float64) error
	// Incr is just Count of 1
	Incr(ctx context.Context, name string, tags []string, rate float64) error
	// Set counts the number of unique elements in a group.
	Set(ctx context.Context, name string, value string, tags []string, rate float64) error
	// Timing sends timing information, it is an alias for TimeInMilliseconds
	Timing(ctx context.Context, name string, value time.Duration, tags []string, rate float64) error
}

var currentBackend = NewNullBackend()

// SetBackend replaces the current backend with the given Backend.
func SetBackend(b Backend) {
	currentBackend = b
}

// Gauge measures the value of a metric at a particular time.
func Gauge(ctx context.Context, name string, value float64, tags []string, rate float64) {
	_ = warnIfError(ctx, currentBackend.Gauge(ctx, name, value, tags, rate))
}

// Count tracks how many times something happened per second.
func Count(ctx context.Context, name string, value int64, tags []string, rate float64) {
	_ = warnIfError(ctx, currentBackend.Count(ctx, name, value, tags, rate))
}

// Histogram tracks the statistical distribution of a set of values on each host.
func Histogram(ctx context.Context, name string, value float64, tags []string, rate float64) {
	_ = warnIfError(ctx, currentBackend.Histogram(ctx, name, value, tags, rate))
}

// Distribution tracks the statistical distribution of a set of values across your infrastructure.
func Distribution(ctx context.Context, name string, value float64, tags []string, rate float64) {
	_ = warnIfError(ctx, currentBackend.Distribution(ctx, name, value, tags, rate))
}

// Decr is just Count of -1
func Decr(ctx context.Context, name string, tags []string, rate float64) {
	_ = warnIfError(ctx, currentBackend.Decr(ctx, name, tags, rate))
}

// Incr is just Count of 1
func Incr(ctx context.Context, name string, tags []string, rate float64) {
	_ = warnIfError(ctx, currentBackend.Incr(ctx, name, tags, rate))
}

// Set counts the number of unique elements in a group.
func Set(ctx context.Context, name string, value string, tags []string, rate float64) {
	_ = warnIfError(ctx, currentBackend.Set(ctx, name, value, tags, rate))
}

// Timing sends timing information, it is an alias for TimeInMilliseconds
func Timing(ctx context.Context, name string, value time.Duration, tags []string, rate float64) {
	_ = warnIfError(ctx, currentBackend.Timing(ctx, name, value, tags, rate))
}

func warnIfError(ctx context.Context, err error) error {
	if err != nil {
		log(ctx, err).WithField("error", err).Warn("couldn't submit event to statsd")
	}
	return err
}

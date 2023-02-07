// Package statsd contains a singleton statsd client for use in all other
// packages. It can be configured once at application startup, and imported by
// any package that wishes to record metrics.
package statsd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/Shopify/goose/v2/logger"
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

// ErrUnknownBackend is returned when the statsd backend implementation is not known.
var ErrUnknownBackend = fmt.Errorf("unknown statsd backend type")

// NewBackend returns the appropriate Backend for the given implementation and host.
// STATSD_DEFAULT_TAGS env variable will be read automatically and added to default tags.
func NewBackend(impl, addr, prefix string, tags ...string) (Backend, error) {
	var err error
	var b Backend
	switch strings.ToLower(impl) {
	case "datadog":
		b, err = NewDatadogBackend(addr, prefix, tags)
	case "log":
		b = NewLogBackend(prefix, tags)
	case "null":
		b = NewNullBackend()
	default:
		return nil, errors.WithStack(ErrUnknownBackend)
	}

	return b, errors.WithStack(err)
}

func warnIfError(ctx context.Context, err error) {
	if err != nil {
		log(ctx, err).WithField("error", err).Warn("couldn't submit event to statsd")
	}
}

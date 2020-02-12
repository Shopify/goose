package statsd

import (
	"context"
	"fmt"
	"strings"
)

// Collector represents a metric that can be collected. It knows about the
// metric name and sampling rate, and supports a Collect method to submit a
// metric to statsd.
type Collector struct {
	Name string
	Rate sampleRate // 0 (default value) is interpreted as 100% (1.0)
}

type sampleRate float64

func (s sampleRate) Rate() float64 {
	if s == 0 {
		return 1.0
	}
	return float64(s)
}

// DefaultRate describes the sample rate used for submitting metrics to StatsD.
const DefaultRate = 0.5

// New returns a new metric with the specified values
func New(name string, rate float64, tags ...string) Metric {
	return Metric{name, rate, tags}
}

// Metric represents a particular metric.
type Metric struct {
	name string
	rate float64
	tags []string
}

func (m Metric) String() string {
	return fmt.Sprintf("%s:%.2f|%s", m.name, m.rate, strings.Join(m.tags, ","))
}

// WithRate returns a new metric based on the current one, only with the rate updated.
func (m Metric) WithRate(rate float64) Metric {
	return Metric{
		name: m.name,
		rate: rate,
		tags: m.tags,
	}
}

// WithTags returns a new metric based on the current one, only with the tags replaced.
func (m Metric) WithTags(tags ...string) Metric {
	return Metric{
		name: m.name,
		rate: m.rate,
		tags: tags,
	}
}

// Count increments a counter based on the metric's name, tags, and rate by the supplied value.
func (m Metric) Count(value int64) {
	Count(context.Background(), m.name, value, m.tags, m.rate)
}

// Distribution tracks the distribution of a set of values based on the metric's name, tags, and rate.
func (m Metric) Distribution(value float64) {
	Distribution(context.Background(), m.name, value, m.tags, m.rate)
}

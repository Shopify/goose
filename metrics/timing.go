package metrics

import (
	"context"
	"time"
)

// Timing represents a timing-type metric, which takes durations and includes percentiles, means, and other information
// along with the event.
// https://github.com/statsd/statsd/blob/master/docs/metric_types.md#timing
type Timing collector

// Duration takes a time.Duration  -- the time the operation took -- and submits it to StatsD.
func (t *Timing) Duration(ctx context.Context, n time.Duration, ts ...Tags) {
	tags := getStatsTags(ctx, ts...)
	warnIfError(ctx, currentBackend.Timing(ctx, t.Name, n, tags, t.Rate.Rate()))
}

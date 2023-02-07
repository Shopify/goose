package metrics

import (
	"context"
)

// Counter represents a count-type metric, which takes increments.
// https://docs.datadoghq.com/developers/metrics/counts/
type Counter collector

// Count takes an integer -- typically 1 -- and increments the counter by
// this value.
//
// The last parameter is an arbitrary array of tags as maps.
func (c *Counter) Count(ctx context.Context, n int64, ts ...Tags) {
	tags := MergeTagsList(ts...)
	backend := BackendFromContext(ctx)
	logError(ctx, backend.Count(ctx, c.Name, n, tags, c.Rate.Rate()))
}

// Incr is basically the same as Count(1)
func (c *Counter) Incr(ctx context.Context, ts ...Tags) {
	c.Count(ctx, 1, ts...)
}

// Decr is basically the same as Count(-1)
func (c *Counter) Decr(ctx context.Context, ts ...Tags) {
	c.Count(ctx, -1, ts...)
}

// SuccessCount is the same as calling Count but adds a `success` tag.
// `success` tag is a boolean based on whether errp points to a nil pointer or not.
func (c *Counter) SuccessCount(ctx context.Context, n int64, errp *error, ts ...Tags) {
	if errp != nil {
		ts = append(ts, Tags{"success": *errp == nil})
	}
	c.Count(ctx, n, ts...)
}

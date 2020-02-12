package statsd

import "context"

// SetCounter represents a set-type metric, which counts unique strings.
// https://docs.datadoghq.com/developers/metrics/sets/
type SetCounter Collector

// CountUnique will count the number of unique elements in a group.
//
// The last parameter is an arbitrary array of tags as maps.
func (c *SetCounter) CountUnique(ctx context.Context, value string, ts ...Tags) {
	tags := getStatsTags(ctx, ts...)
	Set(ctx, c.Name, value, tags, c.Rate.Rate())
}

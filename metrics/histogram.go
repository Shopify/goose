package metrics

import "context"

// Histogram tracks the statistical distribution of a set of values on each host.
// https://docs.datadoghq.com/developers/metrics/types/?tab=histogram#metric-type-definition
type Histogram collector

// The last parameter is an arbitrary array of tags as maps.
func (m *Histogram) Histogram(ctx context.Context, n float64, ts ...Tags) {
	tags := getStatsTagsMap(ctx).Merge(ts...)
	warnIfError(ctx, currentBackend.Histogram(ctx, m.Name, n, tags, m.Rate.Rate()))
}

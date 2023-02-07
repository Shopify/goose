package metrics

import "context"

// Distribution tracks the statistical distribution of a set of values across your infrastructure.
// https://docs.datadoghq.com/developers/metrics/types/?tab=distribution#metric-type-definition
type Distribution collector

// The last parameter is an arbitrary array of tags as maps.
func (d *Distribution) Distribution(ctx context.Context, n float64, ts ...Tags) {
	tags := getStatsTagsMap(ctx).Merge(ts...)
	warnIfError(ctx, currentBackend.Distribution(ctx, d.Name, n, tags, d.Rate.Rate()))
}

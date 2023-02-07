package metrics

import (
	"context"
)

// Distribution tracks the statistical distribution of a set of values across your infrastructure.
// https://docs.datadoghq.com/developers/metrics/types/?tab=distribution#metric-type-definition
type Distribution collector

func (d *Distribution) Distribution(ctx context.Context, n float64, ts ...Tags) {
	tags := MergeTagsList(ts...)
	backend := BackendFromContext(ctx)
	logError(ctx, backend.Distribution(ctx, d.Name, n, tags, d.Rate.Rate()))
}

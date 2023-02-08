package metrics

import (
	"context"
)

// Histogram tracks the statistical distribution of a set of values on each host.
// https://docs.datadoghq.com/developers/metrics/types/?tab=histogram#metric-type-definition
type Histogram collector

func (h *Histogram) Histogram(ctx context.Context, n float64, ts ...Tags) {
	tags := MergeTagsList(ts...)
	backend := BackendFromContext(ctx)
	logError(ctx, backend.Histogram(ctx, h.Name, n, tags, h.Rate.Rate()))
}

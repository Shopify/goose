package metrics

import (
	"context"
)

// NewNullBackend returns a new backend that no-ops every metric.
func NewNullBackend() Backend {
	return NewForwardingBackend(func(_ context.Context, _ *Metric) error {
		return nil
	})
}

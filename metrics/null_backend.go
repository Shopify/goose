package metrics

import (
	"context"
)

// NewNullBackend returns a new backend that no-ops every metric.
func NewNullBackend() Backend {
	return NewForwardingBackend(func(_ context.Context, _ string, _ string, _ interface{}, _ []string, _ float64) error {
		return nil
	})
}

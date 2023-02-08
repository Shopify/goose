package metrics

import (
	"context"
	"sync"
)

// NewRecordingBackend creates a new Backend that records all calls. Useful for tests
func NewRecordingBackend() *RecordingBackend {
	c := &RecordingBackend{}
	c.Backend = NewForwardingBackend(c.record)
	return c
}

var _ Backend = (*RecordingBackend)(nil)

type RecordingBackend struct {
	Backend
	Metrics []*Metric
	lock    sync.Mutex
}

func (c *RecordingBackend) record(ctx context.Context, m *Metric) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.Metrics = append(c.Metrics, m)
	return nil
}

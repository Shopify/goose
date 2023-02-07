package metrics

import (
	"context"

	"github.com/stretchr/testify/mock"
)

var _ Backend = (*MockBackend)(nil)

func NewMockBackend() *MockBackend {
	c := &MockBackend{}
	c.Backend = NewForwardingBackend(c.methodCalled)
	return c
}

type MockBackend struct {
	Backend
	mock.Mock
}

func (c *MockBackend) methodCalled(ctx context.Context, m *Metric) error {
	return c.MethodCalled(m.Type, ctx, m.Name, m.Value, m.Tags, m.Rate).Error(0)
}

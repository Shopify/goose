package concurrency

import (
	"context"

	"github.com/stretchr/testify/mock"
)

var _ Throttler = (*mockThrottler)(nil)

func NewMockThrottler(run bool) *mockThrottler {
	return &mockThrottler{run: run}
}

type mockThrottler struct {
	mock.Mock
	run bool
}

func (m *mockThrottler) Run(ctx context.Context, fn func() error) error {
	args := m.Called(ctx, fn)
	if err := args.Error(0); err != nil {
		return err
	}
	if m.run {
		return fn()
	}
	return nil
}

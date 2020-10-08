package concurrency

import (
	"context"

	"github.com/stretchr/testify/mock"
)

var _ Limiter = &mockLimiter{}

func NewMockLimiter(stubRun bool) *mockLimiter {
	return &mockLimiter{stubRun: stubRun}
}

type mockLimiter struct {
	mock.Mock
	stubRun bool
}

func (l *mockLimiter) MaxConcurrency() uint {
	return l.Called().Get(0).(uint)
}

func (l *mockLimiter) Waiting() int32 {
	return l.Called().Get(0).(int32)
}

func (l *mockLimiter) Running() int32 {
	return l.Called().Get(0).(int32)
}

func (l *mockLimiter) Run(ctx context.Context, fn func() error) error {
	if l.stubRun {
		return fn()
	}
	return l.Called(ctx, fn).Error(0)
}

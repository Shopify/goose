package timetracker

import (
	"time"

	"github.com/stretchr/testify/mock"
)

func NewMockTracker(stubStart bool) *mockTracker {
	return &mockTracker{stubStart: stubStart}
}

var _ Tracker = &mockTracker{}

type mockTracker struct {
	mock.Mock
	stubStart bool
}

func (m *mockTracker) Start() Finisher {
	if m.stubStart {
		return func() {}
	}
	args := m.Called()
	return args.Get(0).(Finisher)
}

func (m *mockTracker) Record(duration time.Duration) {
	m.Called(duration)
}

func (m *mockTracker) Average() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

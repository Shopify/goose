package timetracker

import (
	"testing"
	"time"

	"github.com/leononame/clock"
	"github.com/stretchr/testify/require"
)

func TestExponentialMovingAverageTracker_Record(t *testing.T) {
	tracker := NewExponentialMovingAverageTracker(10)

	require.Equal(t, time.Duration(0), tracker.Average())

	tracker.Record(10000 * time.Millisecond)
	require.Equal(t, 10000*time.Millisecond, tracker.Average())

	tracker.Record(20000 * time.Millisecond)
	require.InDelta(t, 11818*time.Millisecond, tracker.Average(), float64(1*time.Millisecond))

	tracker.Record(20000 * time.Millisecond)
	require.InDelta(t, 13306*time.Millisecond, tracker.Average(), float64(1*time.Millisecond))

	tracker.Record(100000 * time.Millisecond)
	require.InDelta(t, 29068*time.Millisecond, tracker.Average(), float64(1*time.Millisecond))
}

func TestExponentialMovingAverageTracker_Start_End(t *testing.T) {
	tracker := NewExponentialMovingAverageTracker(10).(*exponentialMovingAverageTracker)
	now := time.Now()

	mockClock := clock.NewMock()
	mockClock.Set(now)
	tracker.clock = mockClock

	fn := func() {
		defer tracker.Start().Finish()
		mockClock.Forward(10 * time.Millisecond)
	}

	fn()

	require.Equal(t, 10*time.Millisecond, tracker.Average())
}

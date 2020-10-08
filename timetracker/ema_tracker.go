package timetracker

import (
	"sync"
	"time"

	"github.com/leononame/clock"
)

func NewExponentialMovingAverageTracker(window uint) Tracker {
	return &exponentialMovingAverageTracker{window: window, clock: clock.New()}
}

type exponentialMovingAverageTracker struct {
	clock   clock.Clock
	window  uint
	average int64
	locker  sync.RWMutex
}

func (t *exponentialMovingAverageTracker) Start() Finisher {
	start := t.clock.Now()
	return func() {
		t.Record(t.clock.Since(start))
	}
}

func (t *exponentialMovingAverageTracker) Record(duration time.Duration) {
	t.locker.Lock()
	defer t.locker.Unlock()

	if t.average == 0 {
		t.average = int64(duration)
		return
	}

	// https://www.investopedia.com/ask/answers/122314/what-exponential-moving-average-ema-formula-and-how-ema-calculated.asp
	k := float64(2) / float64(t.window+1)
	previousPortion := float64(t.average) * (1 - k)
	newPortion := float64(duration) * k
	t.average = int64(previousPortion + newPortion)
}

func (t *exponentialMovingAverageTracker) Average() time.Duration {
	t.locker.RLock()
	defer t.locker.RUnlock()

	return time.Duration(t.average)
}

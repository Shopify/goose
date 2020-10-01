package timetracker

import (
	"time"
)

type Tracker interface {
	Start() Finisher
	Record(duration time.Duration)
	Average() time.Duration
}

type Finisher func()

func (f Finisher) Finish() {
	f()
}

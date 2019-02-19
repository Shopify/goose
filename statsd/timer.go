package statsd

import (
	"context"
	"time"
)

// Timer represents a timer-type metric, which takes durations.
// It uses histograms, which is a datadog-specific extension.
// https://docs.datadoghq.com/developers/metrics/histograms/
type Timer Collector

// Duration takes a time.Duration -- the time to complete the indicated
// operation -- and submits it to statsd.
//
// The last parameter is an arbitrary array of tags as maps.
func (t *Timer) Duration(ctx context.Context, n time.Duration, ts ...Tags) {
	tags := loadTags(ctx, t.Tags, ts...)
	Distribution(ctx, t.Name, n.Seconds()*1000, tags, t.Rate.Rate())
}

// Time runs a function, timing its execution, and submits the resulting
// duration to statsd.
//
// The last parameter is an arbitrary array of tags as maps.
func (t *Timer) Time(ctx context.Context, fn func() error, ts ...Tags) error {
	t1 := time.Now()
	err := fn()
	n := time.Since(t1)

	ts = append(ts, Tags{"success": err == nil})
	t.Duration(ctx, n, ts...)
	return err
}

type SuccessFinisher interface {
	SuccessFinish(err *error)
}

type Finisher interface {
	SuccessFinisher
	Finish()
}

type timerFinisher struct {
	timer     *Timer
	startTime time.Time
	tags      []Tags
	ctx       context.Context
}

func (t *timerFinisher) Finish() {
	tags := append([]Tags{t.timer.Tags}, t.tags...)
	t.timer.Duration(t.ctx, time.Since(t.startTime), tags...)
}

func (t *timerFinisher) SuccessFinish(errp *error) {
	tags := append([]Tags{t.timer.Tags}, t.tags...)
	if errp != nil {
		tags = append(tags, Tags{"success": *errp == nil})
	}
	t.timer.Duration(t.ctx, time.Since(t.startTime), tags...)
}

// StartTimer provides a way to collect a duration metric for a function call
// in one line.
//
// Example:
//
//     func foo()  {
//       defer t.StartTimer().Finish()
//       // ...
//     }
func (t *Timer) StartTimer(ctx context.Context, ts ...Tags) Finisher {
	return &timerFinisher{
		timer:     t,
		startTime: time.Now(),
		tags:      ts,
		ctx:       ctx,
	}
}

type successTimerSetFinisher struct {
	set       *SuccessTimerSet
	startTime time.Time
	tags      []Tags
	ctx       context.Context
}

func (s *successTimerSetFinisher) SuccessFinish(errp *error) {
	tags := append([]Tags{s.set.Tags}, s.tags...)
	// Caller should always pass a pointer to an error, but don't crash if it doesn't
	if errp != nil {
		success := *errp == nil
		if success {
			s.set.Success.Incr(s.ctx, tags...)
		} else {
			s.set.Failure.Incr(s.ctx, tags...)
		}
		tags = append(tags, Tags{"success": success})
	}
	s.set.Duration.Duration(s.ctx, time.Since(s.startTime), tags...)
}

// StartTimer provides a way to collect duration and success metrics for a
// function call in one line.
//
// Example:
//
//     func foo() (err error) {
//       defer sts.StartTimer().Finish(&err)
//       err = bar()
//       return
//     }
func (s *SuccessTimerSet) StartTimer(ctx context.Context, ts ...Tags) SuccessFinisher {
	return &successTimerSetFinisher{
		set:       s,
		startTime: time.Now(),
		tags:      ts,
		ctx:       ctx,
	}
}

// SuccessTimerSet is a convenience wrapper around paired "success" and
// "failure" metrics for an operation, as well as a "duration" metric
// indicating how long it took to run the operation.
type SuccessTimerSet struct {
	Success  *Counter
	Failure  *Counter
	Duration *Timer
	Tags     Tags
}

// Instrument executes the provided function, timing its execution.
// It collects the Duration metric with the elapsed time, and increments either
// the success of failure metric depending on whether an error was returned
// from the function.
func (s *SuccessTimerSet) Instrument(ctx context.Context, fn func() error, ts ...Tags) error {
	err := s.Duration.Time(ctx, fn, ts...)
	if err == nil {
		s.Success.Incr(ctx, ts...)
	} else {
		s.Failure.Incr(ctx, ts...)
	}
	return err
}

package statsd

import (
	"context"
	"time"
)

// Timer represents a timer-type metric, which takes durations.
// It uses histograms, which is a datadog-specific extension.
// https://docs.datadoghq.com/developers/metrics/histograms/
type Timer collector

// Duration takes a time.Duration -- the time to complete the indicated
// operation -- and submits it to statsd.
//
// The last parameter is an arbitrary array of tags as maps.
func (t *Timer) Duration(ctx context.Context, n time.Duration, ts ...Tags) {
	tags := getStatsTags(ctx, ts...)
	warnIfError(ctx, currentBackend.Distribution(ctx, t.Name, n.Seconds()*1000, tags, t.Rate.Rate()))
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

type Finisher interface {
	Finish()
	SuccessFinish(err *error)
}

type timerFinisher struct {
	timer     *Timer
	startTime time.Time
	tags      []Tags
	ctx       context.Context
}

func (t *timerFinisher) Finish() {
	t.timer.Duration(t.ctx, time.Since(t.startTime), t.tags...)
}

func (t *timerFinisher) SuccessFinish(errp *error) {
	tags := append([]Tags{}, t.tags...) // Copy the slice
	if errp != nil {
		tags = append(tags, Tags{"success": *errp == nil})
	}
	t.timer.Duration(t.ctx, time.Since(t.startTime), tags...)
}

// StartTimer provides a way to collect a duration metric for a function call
// in one line.
//
// Examples:
//
//     func foo()  {
//       defer t.StartTimer().Finish()
//       // ...
//     }
//
//     func foo() (err error)  {
//       defer t.StartTimer().SuccessFinish(&err)
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

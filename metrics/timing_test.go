package metrics

import (
	"context"
	"testing"
	"time"
)

func TestTiming(t *testing.T) {
	defer func() { SetBackend(NewNullBackend()) }()

	ctx := WithTags(context.Background(), Tags{"test": "value"})
	dur := 1 * time.Millisecond
	statsd := new(MockBackend)
	statsd.On("Timing", ctx, "metric", dur, Tags{"test": "value"}, 1.0).Return(nil)

	SetBackend(statsd)
	metric := &Timing{Name: "metric"}
	metric.Duration(ctx, dur)
}

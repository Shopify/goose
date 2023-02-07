package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/Shopify/goose/v2/metrics/mocks"
)

func TestTiming(t *testing.T) {
	defer func() { SetBackend(NewNullBackend()) }()

	ctx := WithTags(context.Background(), Tags{"test": "value"})
	dur := 1 * time.Millisecond
	statsd := new(mocks.Backend)
	statsd.On("Timing", ctx, "metric", dur, []string{"test:value"}, 1.0).Return(nil)

	SetBackend(statsd)
	metric := &Timing{Name: "metric"}
	metric.Duration(ctx, dur)
}

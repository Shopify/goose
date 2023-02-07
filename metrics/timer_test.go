package metrics

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/Shopify/goose/v2/metrics/mocks"
)

func TestTimer_StartTimer(t *testing.T) {
	defer func() { SetBackend(NewNullBackend()) }()

	ctx := WithTags(context.Background(), Tags{"context": "ok"})

	statsd := new(mocks.Backend)
	SetBackend(statsd)
	metric := &Timer{Name: "metric"}

	t.Run("Finish", func(t *testing.T) {
		statsd.On("Distribution", ctx, "metric", mock.Anything, []string{"context:ok", "finish:ok", "starttimer:ok"}, 1.0).Return(nil).Once()

		start := metric.StartTimer(ctx, Tags{"starttimer": "ok"})
		start.Finish(Tags{"finish": "ok"})

		statsd.AssertExpectations(t)
	})

	t.Run("SuccessFinish", func(t *testing.T) {
		statsd.On("Distribution", ctx, "metric", mock.Anything, []string{"context:ok", "starttimer:ok", "success:false", "successfinish:ok"}, 1.0).Return(nil).Once()

		err := io.EOF
		start := metric.StartTimer(ctx, Tags{"starttimer": "ok"})
		start.SuccessFinish(&err, Tags{"successfinish": "ok"})

		statsd.AssertExpectations(t)
	})

	t.Run("SetTags", func(t *testing.T) {
		statsd.On("Distribution", ctx, "metric", mock.Anything, []string{"context:ok", "settags:ok", "starttimer:ok"}, 1.0).Return(nil).Once()

		start := metric.StartTimer(ctx, Tags{"starttimer": "ok"})
		start.SetTags(Tags{"settags": "ok"})
		start.Finish()

		statsd.AssertExpectations(t)
	})
}

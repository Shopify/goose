package metrics

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/mock"
)

func TestTimer_StartTimer(t *testing.T) {
	metric := &Timer{Name: "metric"}

	t.Run("Finish", func(t *testing.T) {
		backend := NewMockBackend()
		ctx := ContextWithBackend(context.Background(), backend)

		backend.On("Timing", ctx, "metric", mock.Anything, Tags{"finish": "ok", "starttimer": "ok"}, 1.0).Return(nil).Once()

		start := metric.StartTimer(ctx, Tags{"starttimer": "ok"})
		start.Finish(Tags{"finish": "ok"})

		backend.AssertExpectations(t)
	})

	t.Run("SuccessFinish", func(t *testing.T) {
		backend := NewMockBackend()
		ctx := ContextWithBackend(context.Background(), backend)

		backend.On("Timing", ctx, "metric", mock.Anything, Tags{"successfinish": "ok", "starttimer": "ok", "success": false}, 1.0).Return(nil).Once()

		err := io.EOF
		start := metric.StartTimer(ctx, Tags{"starttimer": "ok"})
		start.SuccessFinish(&err, Tags{"successfinish": "ok"})

		backend.AssertExpectations(t)
	})

	t.Run("SetTags", func(t *testing.T) {
		backend := NewMockBackend()
		ctx := ContextWithBackend(context.Background(), backend)

		backend.On("Timing", ctx, "metric", mock.Anything, Tags{"starttimer": "ok", "settags": "ok"}, 1.0).Return(nil).Once()

		start := metric.StartTimer(ctx, Tags{"starttimer": "ok"})
		start.SetTags(Tags{"settags": "ok"})
		start.Finish()

		backend.AssertExpectations(t)
	})
}

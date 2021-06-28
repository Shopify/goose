package maintenance

import (
	"context"
	"fmt"
	"time"

	"github.com/Shopify/go-cache/v2"
	"github.com/Shopify/goose/statsd"

	"github.com/Shopify/courier/pkg/errors"
	"github.com/Shopify/courier/pkg/logging"
	"github.com/Shopify/courier/pkg/maintenance/cursor"
	"github.com/Shopify/courier/pkg/metrics"
)

var log = logging.PackageLogger()

type TaskRunner struct {
	taskName          string
	iterator          Iterator
	executor          Executor
	iterationSlowDown time.Duration
	cursor            cursor.Cursor
	restart           bool
}

func NewTaskRunner(taskName string, iterator Iterator, executor Executor, cache cache.Client) *TaskRunner {
	return &TaskRunner{
		taskName:          taskName,
		iterator:          iterator,
		executor:          executor,
		iterationSlowDown: time.Minute * 1,
		cursor:            cursor.NewCursor(taskName, cache),
	}
}

// SetRestart configure the auto-restart feature.
// When disabled, the runner will process new elements returned by the iterator when starting, then sleep.
func (r *TaskRunner) SetRestart(restart bool) *TaskRunner {
	r.restart = restart
	return r
}

func (r *TaskRunner) Run(ctx context.Context) error {
	ctx = statsd.WithTagLogFields(ctx, statsd.Tags{"task_name": r.taskName})

	for {
		index, err := r.cursor.Current(ctx)
		if err != nil {
			return fmt.Errorf("failed to read cursor: %w", err)
		}

		metrics.TaskIterationProgress.Gauge(ctx, float64(index))

		log(ctx, nil).Infof("loading shops starting at %d", index)

		iterables, nextIndex, err := r.iterator.Next(ctx, int64(index))
		if err != nil {
			return fmt.Errorf("fetching next tasks: %w", err)
		}
		if len(iterables) == 0 {
			log(ctx, nil).Infof("maintenance complete")
			metrics.TaskIterations.Incr(ctx)

			if r.restart {
				if err = r.cursor.Set(ctx, 0); err != nil {
					return errors.Wrap(err, "resetting cursor")
				}

				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(r.iterationSlowDown):
				}

				continue
			}

			<-ctx.Done() // just wait here until cancellation
			return ctx.Err()
		}

		batchStartTime := time.Now()
		if err := r.executor.Perform(ctx, iterables); err != nil {
			return errors.WrapCtx(ctx, err, "maintenance interrupted")
		}
		metrics.TaskBatchExecution.Duration(ctx, time.Since(batchStartTime))

		index = int(nextIndex)

		if err = r.cursor.Set(ctx, index); err != nil {
			return fmt.Errorf("failed to set cursor: %w", err)
		}
	}
}

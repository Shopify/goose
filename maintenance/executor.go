package maintenance

import (
	"context"
	"time"

	"github.com/Shopify/goose/logger"

	"github.com/Shopify/courier/pkg/app/runtime"
	"github.com/Shopify/courier/pkg/errors"
	"github.com/Shopify/courier/pkg/metrics"
)

type Executor interface {
	Perform(ctx context.Context, its []interface{}) error
}

func NewSequentialExecutor(task Task) *sequentialExecutor {
	return &sequentialExecutor{
		task:        task,
		taskTimeout: time.Second * 30,
		rateLimiter: &NoLimitRateLimiter{},
	}
}

type sequentialExecutor struct {
	task        Task
	taskTimeout time.Duration
	rateLimiter RateLimiter
}

func (s sequentialExecutor) WithRateLimiter(rl RateLimiter) *sequentialExecutor {
	s.rateLimiter = rl
	return &s
}

func (s *sequentialExecutor) Perform(ctx context.Context, its []interface{}) error {
	for _, it := range its {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := s.rateLimiter.Wait(ctx)
		if err != nil {
			return err
		}

		err = s.callTask(ctx, it)
		if err != nil {
			return errors.WrapCtx(ctx, err, "task failed")
		}
	}

	return nil
}

func (s *sequentialExecutor) callTask(ctx context.Context, it interface{}) (err error) {
	if withField, ok := it.(logger.Loggable); ok {
		ctx = logger.WithLoggable(ctx, withField)
	} else {
		ctx = logger.WithField(ctx, "task_index", it)
	}

	defer metrics.TaskExecution.StartTimer(ctx).SuccessFinish(&err)

	// Isolate the context to run the tasks:
	// General cancellation will interrupt the runner between tasks, but should not randomly interrupt a task.
	// However we still want a timeout on task execution.
	taskCtx, cancel := context.WithTimeout(runtime.BackgroundContextWithValues(ctx), s.taskTimeout)
	defer cancel()

	log(ctx, nil).Info("task running")
	return s.task.Perform(taskCtx, it)
}

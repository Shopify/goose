package maintenance

import (
	"github.com/Shopify/goose/statsd"
)

var (
	TaskBatchExecution    = &statsd.Timer{Name: "task.batch.execution"}
	TaskExecution         = &statsd.Timer{Name: "task.execution"}
	TaskIterations        = &statsd.Counter{Name: "task.iterations"}
	TaskIterationProgress = &statsd.Gaugor{Name: "task.iteration.progress"}
)

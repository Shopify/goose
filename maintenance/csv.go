package maintenance

import (
	"context"
	"encoding/csv"
	"strings"
)

type CSVTask interface {
	Perform(ctx context.Context, cols []string) error
}

func NewCSVTaskWrapper(task CSVTask) Task {
	return &csvTaskWrapper{task}
}

type csvTaskWrapper struct {
	task CSVTask
}

func (c *csvTaskWrapper) Perform(ctx context.Context, it interface{}) error {
	line := it.(string)

	r := csv.NewReader(strings.NewReader(line))
	cols, err := r.Read()
	if err != nil {
		return err
	}
	return c.task.Perform(ctx, cols)
}

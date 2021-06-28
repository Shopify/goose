package maintenance

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type Iterator interface {
	Next(ctx context.Context, cursor int64) ([]interface{}, int64, error)
}

type Task interface {
	Perform(ctx context.Context, it interface{}) error
}

type TaskMock struct {
	mock.Mock
}

func (s *TaskMock) Perform(ctx context.Context, it interface{}) error {
	return s.Called(ctx, it).Error(0)
}

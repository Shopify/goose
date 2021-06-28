package maintenance

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type testStruct struct {
	int
}

func Test_SequentialExecutor(t *testing.T) {
	ctx := context.Background()

	it1, it2, it3 := int64(100), testStruct{2}, "str-id"

	task := &TaskMock{}
	task.On("Perform", mock.Anything, it1).Once().Return(nil)
	task.On("Perform", mock.Anything, it2).Once().Return(nil)
	task.On("Perform", mock.Anything, it3).Once().Return(errors.New("test-error"))

	executor := NewSequentialExecutor(task)
	err := executor.Perform(ctx, []interface{}{it1, it2, it3})
	require.EqualError(t, err, "task failed: test-error")

	task.AssertExpectations(t)
}

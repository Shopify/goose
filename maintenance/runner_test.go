package maintenance

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Shopify/goose/maintenance/cursor"
)

func TestTaskRunner_Restart(t *testing.T) {
	t.Parallel()

	batch := []interface{}{0, 20}

	iterator := &mockIterator{}
	iterator.On("Next", mock.Anything, int64(0)).Return(batch, int64(20), nil).Once()
	iterator.On("Next", mock.Anything, int64(20)).Return(nil, int64(0), nil).Once()
	iterator.On("Next", mock.Anything, int64(0)).Return(nil, int64(0), context.Canceled).Once()

	task := &TaskMock{}
	task.On("Perform", mock.Anything, batch[0]).Return(nil).Once()
	task.On("Perform", mock.Anything, batch[1]).Return(nil).Once()

	cur := cursor.NewMockCursor(0)
	runner := &TaskRunner{"", iterator, NewSequentialExecutor(task), time.Millisecond, cur, true}

	err := runner.Run(context.Background())
	require.EqualError(t, err, "fetching next tasks: context canceled")

	iterator.AssertExpectations(t)
	task.AssertExpectations(t)
	cur.Assert(t, 0) // cursor is resetted on completion
}

func TestTaskRunner_Restart_Disabled(t *testing.T) {
	t.Parallel()

	batch := []interface{}{10, 20}

	iterator := &mockIterator{}
	iterator.On("Next", mock.Anything, int64(0)).Return(batch, int64(20), nil).Once()
	iterator.On("Next", mock.Anything, int64(20)).Return(nil, int64(0), context.Canceled).Once()

	task := &TaskMock{}
	task.On("Perform", mock.Anything, batch[0]).Return(nil).Once()
	task.On("Perform", mock.Anything, batch[1]).Return(nil).Once()

	cur := cursor.NewMockCursor(0)
	runner := &TaskRunner{"", iterator, NewSequentialExecutor(task), time.Second, cur, false}

	err := runner.Run(context.Background())
	require.EqualError(t, err, "fetching next tasks: context canceled")

	iterator.AssertExpectations(t)
	task.AssertExpectations(t)
	cur.Assert(t, 20) // cursor is not reset on completion
}

func TestTaskRunner_RestartFromCursor(t *testing.T) {
	t.Parallel()

	iterator := &mockIterator{}
	iterator.On("Next", mock.Anything, int64(20)).Return(nil, int64(0), context.Canceled).Once()

	runner := &TaskRunner{"", iterator, NewSequentialExecutor(nil), time.Second, cursor.NewMockCursor(20), true}

	err := runner.Run(context.Background())
	require.EqualError(t, err, "fetching next tasks: context canceled")

	iterator.AssertExpectations(t)
}

func TestTaskRunner_TaskError(t *testing.T) {
	t.Parallel()

	batch := []interface{}{10, 20}

	iterator := &mockIterator{}
	iterator.On("Next", mock.Anything, int64(0)).Return(batch, int64(20), nil).Once()

	task := &TaskMock{}
	task.On("Perform", mock.Anything, batch[0]).Return(errors.New("testtest")).Once()

	cur := cursor.NewMockCursor(0)
	runner := &TaskRunner{"", iterator, NewSequentialExecutor(task), time.Second, cur, false}

	err := runner.Run(context.Background())
	require.EqualError(t, err, "maintenance interrupted: task failed: testtest")

	iterator.AssertExpectations(t)
	task.AssertExpectations(t)
	cur.Assert(t, 0) // cursor is not moved if one of the task fails
}

func TestTaskRunner_Cancellation(t *testing.T) {
	t.Parallel()

	batch := []interface{}{10, 20}

	iterator := &mockIterator{}
	iterator.On("Next", mock.Anything, int64(0)).Return(batch, int64(20), nil).Once()

	runner := &TaskRunner{"", iterator, NewSequentialExecutor(nil), time.Second, cursor.NewMockCursor(0), false}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runner.Run(ctx)
	require.True(t, errors.Is(err, context.Canceled))

	iterator.AssertExpectations(t)
}

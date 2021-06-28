package maintenance

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockCSVTask struct {
	mock.Mock
}

func (m *mockCSVTask) Perform(ctx context.Context, cols []string) error {
	args := m.Called(ctx, cols)
	return args.Error(0)
}

func TestCSVTask(t *testing.T) {
	m := &mockCSVTask{}
	task := NewCSVTaskWrapper(m)

	m.On("Perform", mock.Anything, []string{"foo", "1234567890"}).Return(nil).Once()
	err := task.Perform(context.Background(), `"foo",1234567890`)
	require.NoError(t, err)

	m.AssertExpectations(t)
}

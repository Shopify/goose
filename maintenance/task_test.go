package maintenance

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type mockIterator struct {
	mock.Mock
}

var _ Iterator = &mockIterator{}

func (m *mockIterator) Next(ctx context.Context, cursor int64) ([]interface{}, int64, error) {
	args := m.Called(ctx, cursor)
	if args.Get(0) != nil {
		return args.Get(0).([]interface{}), args.Get(1).(int64), args.Error(2)
	}

	return nil, 0, args.Error(2)
}

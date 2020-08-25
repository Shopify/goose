package random

import (
	"math/rand"

	"github.com/stretchr/testify/mock"
)

var _ rand.Source = (*mockSource)(nil)

// NewMockSource returns a mocked rand.Source. Configure each call before using.
func NewMockSource() *mockSource {
	return &mockSource{}
}

type mockSource struct {
	mock.Mock
}

func (s *mockSource) Int63() int64 {
	return s.Called().Get(0).(int64)
}

func (s *mockSource) Seed(seed int64) {
	s.Called(seed)
}

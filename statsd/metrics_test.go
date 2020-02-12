package statsd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Shopify/goose/statsd/mocks"
)

func TestWithRate(t *testing.T) {
	m := New("some_metric", 0.3, "kind:counter")
	m2 := m.WithRate(0.5)

	require.Equal(t, "some_metric:0.30|kind:counter", m.String())
	require.Equal(t, "some_metric:0.50|kind:counter", m2.String())
}

func TestMetricsWithTags(t *testing.T) {
	m := New("some_metric", 0.3, "kind:counter")
	m2 := m.WithTags("kind:gauge")

	require.Equal(t, "some_metric:0.30|kind:counter", m.String())
	require.Equal(t, "some_metric:0.30|kind:gauge", m2.String())
}

func TestCount(t *testing.T) {
	m := New("some_cmd", DefaultRate, "tags")

	backend := new(mocks.Backend)
	backend.On("Count", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	SetBackend(backend)

	m.Count(100)
	backend.AssertCalled(t, "Count", context.Background(), "some_cmd", int64(100), []string{"tags"}, DefaultRate)
}

func TestDistribution(t *testing.T) {
	m := New("thing", 0.3, "t1", "t2")

	backend := new(mocks.Backend)
	backend.On("Distribution", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	SetBackend(backend)

	m.Distribution(10.2)
	backend.AssertCalled(t, "Distribution", context.Background(), "thing", 10.2, []string{"t1", "t2"}, 0.3)
}

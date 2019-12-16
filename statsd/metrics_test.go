package statsd

import (
	"context"
	"errors"
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

	tests := []struct {
		expected []interface{}
		err      error
	}{
		{expected: []interface{}{context.Background(), "some_cmd", int64(100), []string{"tags"}, DefaultRate}},
		{expected: []interface{}{context.Background(), "some_cmd", int64(100), []string{"tags"}, DefaultRate}, err: errors.New("boom")},
	}

	for _, test := range tests {
		backend := new(mocks.Backend)
		backend.On("Count", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(test.err)
		SetBackend(backend)

		err := m.Count(100)
		require.Equal(t, test.err, err)
		backend.AssertCalled(t, "Count", test.expected...)
	}
}

func TestIncr(t *testing.T) {
	m := New("some_cmd", DefaultRate, "tags")

	tests := []struct {
		expected []interface{}
		err      error
	}{
		{expected: []interface{}{context.Background(), "some_cmd", []string{"tags"}, DefaultRate}},
		{expected: []interface{}{context.Background(), "some_cmd", []string{"tags"}, DefaultRate}, err: errors.New("boom")},
	}

	for _, test := range tests {
		backend := new(mocks.Backend)
		backend.On("Incr", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(test.err)
		SetBackend(backend)

		err := m.Incr()
		require.Equal(t, test.err, err)
		backend.AssertCalled(t, "Incr", test.expected...)
	}
}

func TestDecr(t *testing.T) {
	m := New("some_ctr", 0.3, "t1", "t2")

	tests := []struct {
		expected []interface{}
		err      error
	}{
		{expected: []interface{}{context.Background(), "some_ctr", []string{"t1", "t2"}, 0.3}},
		{expected: []interface{}{context.Background(), "some_ctr", []string{"t1", "t2"}, 0.3}, err: errors.New("boom")},
	}

	for _, test := range tests {
		backend := new(mocks.Backend)
		backend.On("Decr", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(test.err)
		SetBackend(backend)

		err := m.Decr()
		require.Equal(t, test.err, err)
		backend.AssertCalled(t, "Decr", test.expected...)
	}
}

func TestDistribution(t *testing.T) {
	m := New("thing", 0.3, "t1", "t2")

	tests := []struct {
		expected []interface{}
		err      error
	}{
		{expected: []interface{}{context.Background(), "thing", 10.2, []string{"t1", "t2"}, 0.3}},
		{expected: []interface{}{context.Background(), "thing", 10.2, []string{"t1", "t2"}, 0.3}, err: errors.New("boom")},
	}

	for _, test := range tests {
		backend := new(mocks.Backend)
		backend.On("Distribution", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(test.err)
		SetBackend(backend)

		err := m.Distribution(10.2)
		require.Equal(t, test.err, err)
		backend.AssertCalled(t, "Distribution", test.expected...)
	}
}

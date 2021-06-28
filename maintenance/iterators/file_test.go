package iterators

import (
	"context"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/Shopify/go-storage"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var dummyFile = `1234567890
0987654321
1234509876
`

func TestFileIterator__ReadsLineAtCursor(t *testing.T) {
	fs := storage.NewMemoryFS()
	err := storage.Write(context.Background(), fs, "foo", []byte(dummyFile), nil)
	require.NoError(t, err)

	iterator := NewFileIterator("foo", fs)

	lines, next, err := iterator.Next(context.Background(), 0)
	require.NoError(t, err)
	require.Equal(t, []interface{}{"1234567890"}, lines)
	require.Equal(t, int64(1), next)

	lines, next, err = iterator.Next(context.Background(), next)
	require.NoError(t, err)
	require.Equal(t, []interface{}{"0987654321"}, lines)
	require.Equal(t, int64(2), next)

	lines, next, err = iterator.Next(context.Background(), next)
	require.NoError(t, err)
	require.Equal(t, []interface{}{"1234509876"}, lines)
	require.Equal(t, int64(3), next)

	lines, next, err = iterator.Next(context.Background(), next)
	require.NoError(t, err)
	require.Nil(t, lines)
	require.Equal(t, int64(0), next)
}

func TestFileIterator__SkipsReadLine(t *testing.T) {
	mockStorage := storage.NewMockFS()

	iterator := NewFileIterator("config/myfile.csv", mockStorage)

	mockStorage.On("Open", mock.Anything, "config/myfile.csv", (*storage.ReaderOptions)(nil)).Return(&storage.File{
		ReadCloser: ioutil.NopCloser(strings.NewReader(dummyFile)),
	}, nil)

	lines, next, err := iterator.Next(context.Background(), 2)
	require.NoError(t, err)
	require.Equal(t, []interface{}{"1234509876"}, lines)
	require.Equal(t, int64(3), next)
}

func TestFileIterator__Fixture(t *testing.T) {
	iterator := NewFileIterator("fixtures/file.txt", storage.NewLocalFS(""))

	lines, next, err := iterator.Next(context.Background(), 0)
	require.NoError(t, err)
	require.Equal(t, []interface{}{"a"}, lines)
	require.Equal(t, int64(1), next)

	lines, next, err = iterator.Next(context.Background(), next)
	require.NoError(t, err)
	require.Equal(t, []interface{}{"b"}, lines)
	require.Equal(t, int64(2), next)

	lines, next, err = iterator.Next(context.Background(), next)
	require.NoError(t, err)
	require.Equal(t, []interface{}{"c"}, lines)
	require.Equal(t, int64(3), next)

	lines, next, err = iterator.Next(context.Background(), next)
	require.NoError(t, err)
	require.Nil(t, lines)
	require.Equal(t, int64(0), next)
}

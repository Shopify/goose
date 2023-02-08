package errors

import (
	"context"
	stderrors "errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	err := New("MSG")
	require.Equal(t, "MSG", err.Error())
}

func TestErrorf(t *testing.T) {
	err := Errorf("MSG %s", "ME")
	require.Equal(t, "MSG ME", err.Error())
}

func TestWrap_Nil(t *testing.T) {
	err := Wrap(nil, "")
	require.Nil(t, err)
}

func TestWrap(t *testing.T) {
	err := Wrap(stderrors.New("inner"), "outer")
	require.NotNil(t, err)
	require.Equal(t, "outer: inner", err.Error())
}

func TestWrapCtx_Nil(t *testing.T) {
	ctx := context.Background()
	err := WrapCtx(ctx, nil, "")
	require.Nil(t, err)
}

func TestWrapCtx(t *testing.T) {
	ctx := context.Background()
	err := WrapCtx(ctx, stderrors.New("inner"), "outer")
	require.NotNil(t, err)
	require.Equal(t, "outer: inner", err.Error())
}

func TestWithCtx(t *testing.T) {
	ctx := context.Background()
	err := WithCtx(ctx, stderrors.New("inner"), Fields{"key": "val"})
	require.NotNil(t, err)
	require.Equal(t, "inner", err.Error())

	require.Equal(t, Fields{"key": "val"}, FieldsFromError(err))
}

func TestWith(t *testing.T) {
	err := With(stderrors.New("inner"), Fields{"key": "val"})
	require.NotNil(t, err)
	require.Equal(t, "inner", err.Error())

	require.Equal(t, Fields{"key": "val"}, FieldsFromError(err))
}

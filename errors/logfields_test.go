package errors

import (
	"context"
	stderrors "errors"
	"testing"

	"github.com/Shopify/goose/v2/logger"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func Test_FieldsFromError_NoFields(t *testing.T) {
	require.Empty(t, FieldsFromError(nil))
	require.Empty(t, FieldsFromError(New("")))
}

type testError struct {
	error
	fields logrus.Fields
}

func (e *testError) Unwrap() error {
	return e.error
}

func (e *testError) LogFields() logrus.Fields {
	return e.fields
}

func Test_FieldsFromError(t *testing.T) {
	err1 := &testError{error: New("foo"), fields: logrus.Fields{"KEY1": "VAL1", "KEY2": "VAL2"}}
	err2 := Wrap(err1, "bar", Fields{"KEY2": "VAL3", "KEY3": "VAL3"})

	// Fields from outer error have precedence
	require.Equal(t, Fields{"KEY1": "VAL1", "KEY2": "VAL3", "KEY3": "VAL3"}, FieldsFromError(err2))
}

func Test_FieldsFromError_From_Context(t *testing.T) {
	originalErr := Wrap(New(""), "", Fields{"KEY1": "VAL1"})

	ctx := context.Background()
	ctx = logger.WithFields(ctx, logrus.Fields{"KEY1": "VAL1", "KEY2": "VAL2"})

	err := WrapCtx(ctx, originalErr, "", Fields{"KEY2": "EXTRA", "EXTRA": "EXTRA"}) // KEY2 overlap
	require.Equal(t, Fields{
		"KEY1":  "VAL1",
		"KEY2":  "VAL2", // fields from inner error have precedence.
		"EXTRA": "EXTRA",
	}, FieldsFromError(err))
}

func Test_FieldsFromJoinedError(t *testing.T) {
	err1 := Wrap(New(""), "", Fields{"FOO": "BAR"})
	err2 := stderrors.Join(Wrap(err1, "", Fields{"BAZ": "BOO"}), New("second"))

	extracted := FieldsFromError(err2)
	require.Equal(t, Fields{"FOO": "BAR", "BAZ": "BOO"}, extracted)

	err3 := stderrors.Join(Wrap(err1, "", Fields{"BAZ": "BOO"}), Wrap(New(""), "", Fields{"FOO": "BAR"}))

	extracted = FieldsFromError(err3)
	require.Equal(t, Fields{"FOO": "BAR", "BAZ": "BOO"}, extracted)

	err4 := stderrors.Join(Wrap(New(""), "", Fields{"FOO": "BAR"}), Wrap(New(""), "", Fields{"BAZ": "BOO"}))

	extracted = FieldsFromError(err4)
	require.Equal(t, Fields{"FOO": "BAR", "BAZ": "BOO"}, extracted)

	err5 := stderrors.Join(Wrap(New(""), "", Fields{"FRUIT": "BANANA"}), New(""))
	err6 := Wrap(stderrors.Join(Wrap(err5, "", Fields{"BAZ": "BOO"}), Wrap(New(""), "", Fields{"FOO": "BAR"})), "", Fields{"JOINED": "YES"})

	extracted = FieldsFromError(err6)
	require.Equal(t, Fields{"FOO": "BAR", "BAZ": "BOO", "JOINED": "YES", "FRUIT": "BANANA"}, extracted)
}

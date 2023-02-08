package errors

import (
	"context"
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

package bugsnag

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestExtractWithoutErrorClass(t *testing.T) {
	err := errors.New("test error")
	event := captureNotifyEvent(t, err)

	require.EqualError(t, event.Error, "test error")
	require.Equal(t, "basic_error", event.MetaData["error"]["type"])
	require.Nil(t, event.MetaData["error"]["class"])
}

func TestWithErrorClass(t *testing.T) {
	err := WithErrorClass(errors.New("test error"), "FOO")
	event := captureNotifyEvent(t, err)

	require.EqualError(t, event.Error, "test error")
	require.Equal(t, "basic_error", event.MetaData["error"]["type"])
	require.Equal(t, "FOO", event.ErrorClass)
}

func TestWrappedWithErrorClass(t *testing.T) {
	err := errors.Wrapf(WithErrorClass(errors.New("test error"), "FOO"), "other %v", "bar")
	event := captureNotifyEvent(t, err)

	require.EqualError(t, event.Error, "other bar: test error")
	require.Equal(t, "basic_error", event.MetaData["error"]["type"])
	require.Equal(t, "FOO", event.ErrorClass)
}

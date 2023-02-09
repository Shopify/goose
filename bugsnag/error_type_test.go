package bugsnag

import (
	"errors"
	"testing"

	pkgerrors "github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestBasicError(t *testing.T) {
	event := captureNotifyEvent(t, errors.New("foo"))

	require.EqualError(t, event.Error, "foo")
	require.Equal(t, "basic_error", event.MetaData["error"]["type"])
}

func TestPkgError(t *testing.T) {
	err := pkgerrors.New("foo")
	event := captureNotifyEvent(t, err)

	require.EqualError(t, event.Error, "foo")
	require.Equal(t, "basic_error", event.MetaData["error"]["type"])
}

func TestErrorWrap(t *testing.T) {
	err := pkgerrors.Wrap(pkgerrors.Wrap(errors.New("foo"), "bar"), "baz")
	event := captureNotifyEvent(t, err)

	require.EqualError(t, event.Error, "baz: bar: foo")
	require.Equal(t, "basic_error", event.MetaData["error"]["type"])
}

func TestErrorWrapTyped(t *testing.T) {
	err := pkgerrors.Wrap(testError("foo"), "bar")
	event := captureNotifyEvent(t, err)

	require.EqualError(t, event.Error, "bar: foo")
	require.Equal(t, "bugsnag.testError", event.MetaData["error"]["type"])
}

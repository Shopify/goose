package bugsnag

import (
	"testing"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestLogErrorHook(t *testing.T) {
	event := captureEvent(t, func() {
		err := errors.Wrap(errors.New("baz"), "bar")

		log.WithFields(log.Fields{
			"key": "val",
			"int": 12345,
		}).WithError(err).Error("foo")
	})

	require.EqualError(t, event.Error, "bar: baz")
	require.Equal(t, "basic_error", event.MetaData["error"]["type"])
	require.Equal(t, "", event.Context)
	require.Equal(t, "val", event.MetaData["metadata"]["key"])
	require.Equal(t, 12345, event.MetaData["metadata"]["int"])
}

func TestLogEntry(t *testing.T) {
	err := errors.Wrap(errors.New("baz"), "bar")
	entry := log.WithFields(log.Fields{
		"key": "val",
		"int": 12345,
	}).WithError(err)
	entry.Message = "foo"

	event := captureNotifyEvent(t, err, entry)

	require.EqualError(t, event.Error, "bar: baz")
	require.Equal(t, "basic_error", event.MetaData["error"]["type"])
	require.Equal(t, "foo", event.Context)
	require.Equal(t, "foo", event.MetaData["log"]["message"])
	require.Equal(t, "val", event.MetaData["metadata"]["key"])
	require.Equal(t, 12345, event.MetaData["metadata"]["int"])
}

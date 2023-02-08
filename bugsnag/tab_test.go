package bugsnag

import (
	"testing"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/require"
)

func TestTab(t *testing.T) {
	err := errors.Wrap(errors.New("baz"), "bar")
	tab := Tab{
		Label: "test",
		Rows: map[string]any{
			"err": 1,
		},
	}
	event := captureNotifyEvent(t, err, tab)

	require.EqualError(t, event.Error, "bar: baz")
	require.Equal(t, "", event.Context)
	require.Equal(t, 1, event.MetaData["test"]["err"])
}

type tabProvider struct {
	val string
}

func (ct *tabProvider) CreateBugsnagTab() Tab {
	return Tab{
		Label: "custom",
		Rows:  Rows{"key": ct.val},
	}
}

func TestTabProvider(t *testing.T) {
	err := errors.Wrap(errors.New("baz"), "bar")
	event := captureNotifyEvent(t, err, &tabProvider{"val"})

	require.EqualError(t, event.Error, "bar: baz")
	require.Equal(t, "val", event.MetaData["custom"]["key"])
}

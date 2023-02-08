package bugsnag

import (
	"flag"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	bugsnaggo "github.com/bugsnag/bugsnag-go/v2"
)

type testError string

func (e testError) Error() string {
	return string(e)
}

func TestMain(m *testing.M) {
	if !flag.Parsed() {
		flag.Parse()
	}

	// This test is incompatible with paniconexit0
	panicFlag := flag.CommandLine.Lookup("test.paniconexit0")
	if panicFlag != nil {
		err := panicFlag.Value.Set("false")
		if err != nil {
			panic(err)
		}
	}

	AutoConfigure("api-key", "api-version", "service-test", []string{})
	code := m.Run()
	os.Exit(code)
}

func captureEvent(t *testing.T, fn func()) *bugsnaggo.Event {
	events := GlobalHook.CaptureEvents(fn)
	require.Len(t, events, 1)
	return events[0]
}

func captureNotifyEvent(t *testing.T, err error, rawData ...any) *bugsnaggo.Event {
	return captureEvent(t, func() {
		notifyErr := bugsnaggo.Notify(err, rawData...)
		require.ErrorIs(t, notifyErr, ErrCaptured)
	})
}

func TestHttpRequest(t *testing.T) {
	req := &http.Request{
		Method: "GET",
		Host:   "example.com",
		URL: &url.URL{
			Path: "/example",
		},
	}
	err := testError("foo")
	event := captureNotifyEvent(t, err, req)

	require.EqualError(t, event.Error, "foo")
	require.Equal(t, "bugsnag.testError", event.MetaData["error"]["type"])
	require.Equal(t, "http://example.com/example", event.Request.URL)
	require.Equal(t, "/example", event.Context)
}

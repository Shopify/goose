package logrusbugsnag

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"sync"
	"testing"

	"github.com/bugsnag/bugsnag-go/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Copied from bugsnag tests

var roundTripper = &nilRoundTripper{}
var events = make(chan *bugsnag.Event, 10)
var testOnce sync.Once
var testAPIKey = "166f5ad3590596f9aa8d601ea89af845"
var errTest = errors.New("test error")

type nilRoundTripper struct{}

func (rt *nilRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	return &http.Response{
		Body:       io.NopCloser(bytes.NewReader(nil)),
		StatusCode: http.StatusOK,
	}, nil
}

func setup(record bool) {
	testOnce.Do(func() {
		l := logrus.New()
		l.Out = io.Discard

		bugsnag.Configure(bugsnag.Configuration{
			APIKey: testAPIKey,
			Endpoints: bugsnag.Endpoints{
				Notify: "",
			},
			Synchronous: true,
			Transport:   roundTripper,
			Logger:      l,
		})
		if record {
			bugsnag.OnBeforeNotify(func(event *bugsnag.Event, config *bugsnag.Configuration) error {
				events <- event
				return nil
			})
		}
	})
}

func BenchmarkHook_Fire(b *testing.B) {
	setup(false)

	l := logrus.New()
	l.Out = io.Discard

	hook, err := NewBugsnagHook(nil)
	assert.NoError(b, err)
	l.Hooks.Add(hook)

	b.Run("Error", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			l.Error("testing")
		}
	})

	var doRecover = func() {
		recover()
	}

	var doPanic = func() {
		defer doRecover()
		l.Panic("testing")
	}

	b.Run("Panic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			doPanic()
		}
	})
}

func TestNewBugsnagHook(t *testing.T) {
	setup(true)

	l := logrus.New()
	l.Out = io.Discard

	hook, err := NewBugsnagHook(nil)
	assert.NoError(t, err)
	l.Hooks.Add(hook)

	t.Run("inline error", func(t *testing.T) {
		t.Run("inline logging", func(t *testing.T) {
			l.WithError(err).Error(errors.New("foo"))

			event := <-events
			assert.Equal(t, "*errors.errorString", event.ErrorClass)
			assert.Equal(t, "foo", event.Message)
			assert.NotEqual(t, "triggerError", event.Stacktrace[0].Method)
			assert.Contains(t, event.Stacktrace[0].File, "bugsnag_test.go")
		})

		t.Run("other function logging", func(t *testing.T) {
			triggerError(l, errors.New("foo"))

			event := <-events
			assert.Equal(t, "*errors.errorString", event.ErrorClass)
			assert.Equal(t, "foo", event.Message)
			assert.Equal(t, "triggerError", event.Stacktrace[0].Method)
		})
	})

	t.Run("prebuilt error", func(t *testing.T) {
		t.Run("inline logging", func(t *testing.T) {
			l.WithError(errTest).Error("test")

			event := <-events
			assert.Equal(t, "*errors.errorString", event.ErrorClass)
			assert.Equal(t, "test error", event.Message)
			assert.NotEqual(t, "triggerError", event.Stacktrace[0].Method)
			assert.Contains(t, event.Stacktrace[0].File, "bugsnag_test.go")
		})

		t.Run("other function logging", func(t *testing.T) {
			triggerError(l, errTest)

			event := <-events
			assert.Equal(t, "*errors.errorString", event.ErrorClass)
			assert.Equal(t, "test error", event.Message)
			assert.Equal(t, "triggerError", event.Stacktrace[0].Method)
		})
	})
}

func triggerError(l *logrus.Logger, err error) {
	l.WithError(err).Error("test")
}

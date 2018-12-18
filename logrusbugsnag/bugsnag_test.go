package logrusbugsnag

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"

	"github.com/bitly/go-simplejson"
	"github.com/bugsnag/bugsnag-go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Copied from bugsnag tests

var roundTripper = &nilRoundTripper{}
var postedJSON = make(chan []byte, 10)
var testOnce sync.Once
var testAPIKey = "166f5ad3590596f9aa8d601ea89af845"
var errTest = errors.New("test error")

type nilRoundTripper struct {
	record bool
}

func (rt *nilRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if rt.record {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		postedJSON <- body
	}

	return &http.Response{
		Body:       ioutil.NopCloser(bytes.NewReader(nil)),
		StatusCode: http.StatusOK,
	}, nil
}

func setup(record bool) {
	roundTripper.record = record
	testOnce.Do(func() {
		l := logrus.New()
		l.Out = ioutil.Discard

		bugsnag.Configure(bugsnag.Configuration{
			APIKey: testAPIKey,
			Endpoints: bugsnag.Endpoints{
				Notify: "",
			},
			Synchronous: true,
			Transport:   roundTripper,
			Logger:      l,
		})
	})
}

func BenchmarkHook_Fire(b *testing.B) {
	setup(false)

	l := logrus.New()
	l.Out = ioutil.Discard

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
	l.Out = ioutil.Discard

	hook, err := NewBugsnagHook(nil)
	assert.NoError(t, err)
	l.Hooks.Add(hook)

	t.Run("inline error", func(t *testing.T) {
		t.Run("inline logging", func(t *testing.T) {
			l.WithError(err).Error(errors.New("foo"))
			json, err := simplejson.NewJson(<-postedJSON)
			assert.NoError(t, err)

			exception := json.Get("events").GetIndex(0).Get("exceptions").GetIndex(0)
			assert.Equal(t, "*errors.errorString", exception.Get("errorClass").MustString())
			assert.Equal(t, "foo", exception.Get("message").MustString())
			assert.NotEqual(t, "triggerError", exception.Get("stacktrace").GetIndex(0).Get("method").MustString())
			assert.Contains(t, exception.Get("stacktrace").GetIndex(0).Get("file").MustString(), "bugsnag_test.go")
		})

		t.Run("other function logging", func(t *testing.T) {
			triggerError(l, errors.New("foo"))
			json, err := simplejson.NewJson(<-postedJSON)
			assert.NoError(t, err)

			exception := json.Get("events").GetIndex(0).Get("exceptions").GetIndex(0)
			assert.Equal(t, "*errors.errorString", exception.Get("errorClass").MustString())
			assert.Equal(t, "foo", exception.Get("message").MustString())
			assert.Equal(t, "triggerError", exception.Get("stacktrace").GetIndex(0).Get("method").MustString())
		})
	})

	t.Run("prebuilt error", func(t *testing.T) {
		t.Run("inline logging", func(t *testing.T) {
			l.WithError(errTest).Error("test")
			json, err := simplejson.NewJson(<-postedJSON)
			assert.NoError(t, err)

			exception := json.Get("events").GetIndex(0).Get("exceptions").GetIndex(0)
			assert.Equal(t, "*errors.errorString", exception.Get("errorClass").MustString())
			assert.Equal(t, "test error", exception.Get("message").MustString())
			assert.NotEqual(t, "triggerError", exception.Get("stacktrace").GetIndex(0).Get("method").MustString())
			assert.Contains(t, exception.Get("stacktrace").GetIndex(0).Get("file").MustString(), "bugsnag_test.go")
		})

		t.Run("other function logging", func(t *testing.T) {
			triggerError(l, errTest)
			json, err := simplejson.NewJson(<-postedJSON)
			assert.NoError(t, err)

			exception := json.Get("events").GetIndex(0).Get("exceptions").GetIndex(0)
			assert.Equal(t, "*errors.errorString", exception.Get("errorClass").MustString())
			assert.Equal(t, "test error", exception.Get("message").MustString())
			assert.Equal(t, "triggerError", exception.Get("stacktrace").GetIndex(0).Get("method").MustString())
		})
	})
}

func triggerError(l *logrus.Logger, err error) {
	l.WithError(err).Error("test")
}

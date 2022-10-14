package srvutil

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/tomb.v2"

	"github.com/Shopify/goose/metrics"
	"github.com/Shopify/goose/safely"
	"github.com/Shopify/goose/statsd"
)

func TestRequestMetricsMiddleware(t *testing.T) {
	var recordedTags []string
	statsd.SetBackend(statsd.NewForwardingBackend(func(_ context.Context, mType string, name string, value interface{}, tags []string, _ float64) error {
		if name == metrics.HTTPRequest.Name {
			recordedTags = tags
		}
		return nil
	}))

	logOutput := logrus.StandardLogger().Out
	defer logrus.StandardLogger().SetOutput(logOutput)
	logging := &bytes.Buffer{}
	logrus.StandardLogger().SetOutput(logging)

	logLevel := logrus.StandardLogger().Level
	defer logrus.StandardLogger().SetLevel(logLevel)
	logrus.StandardLogger().SetLevel(logrus.DebugLevel)

	tb := &tomb.Tomb{}
	sl := FuncServlet("/hello/{name}", func(res http.ResponseWriter, req *http.Request) {
		name := mux.Vars(req)["name"]
		res.Header().Set("foo", "bar")
		res.Header().Set("set-cookie", "secret")
		fmt.Fprintf(res, "hello %s", name)
	})

	sl = UseServlet(
		sl,
		RequestContextMiddleware,
		NewRequestMetricsMiddleware(&RequestMetricsMiddlewareConfig{BodyLogPredicate: LogErrorBody}),
	)

	s := NewServer(tb, "127.0.0.1:0", sl)
	defer s.Tomb().Kill(nil)
	safely.Run(s)

	u := "http://" + s.Addr().String() + "/hello/world"

	req, err := http.NewRequest("GET", u, nil)
	req.Header.Set("Authorization", "secret")
	req.Header.Set("Cookie", "secret")
	req.Header.Set("Foo", "baz")
	assert.NoError(t, err)

	// Works
	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "hello world", string(body))

	assert.NotNil(t, recordedTags, "should have recorded a %s tag", metrics.HTTPRequest.Name)
	assert.Equal(t, []string{"route:/hello/@name", "statusClass:2xx", "statusCode:200"}, recordedTags)

	output := strings.ToLower(logging.String())
	assert.Contains(t, output, "statusclass=2xx")
	assert.Contains(t, output, "statuscode=200")
	assert.Contains(t, output, "foo:bar")
	assert.Contains(t, output, "foo:baz")
	assert.NotContains(t, output, "secret")
	assert.Contains(t, output, "authorization:[filtered]")
	assert.Contains(t, output, "cookie:[filtered]")
	assert.Contains(t, output, "set-cookie:[filtered]")
}

type dummyHijackableResponseWriter struct {
}

func (dummyHijackableResponseWriter) Header() http.Header {
	return http.Header{}
}

func (dummyHijackableResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (dummyHijackableResponseWriter) WriteHeader(statusCode int) {

}

func (dummyHijackableResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, nil
}

var _ http.ResponseWriter = &dummyHijackableResponseWriter{}
var _ http.Hijacker = &dummyHijackableResponseWriter{}

func TestNewHTTPRecorder(t *testing.T) {
	t.Run("regular response writer", func(t *testing.T) {
		w := httptest.NewRecorder()
		recorder := newHTTPRecorder(w, nil)

		assert.IsType(t, &httpRecorder{}, recorder)
		_, ok := recorder.(http.ResponseWriter)
		assert.True(t, ok, "recorder must implement http.ResponseWriter")
		_, ok = recorder.(http.Hijacker)
		assert.False(t, ok, "recorder must not implement http.Hijacker")

		recorder.Header().Set("foo", "bar")
		recorder.WriteHeader(http.StatusAccepted)
		_, err := recorder.Write([]byte("the body"))
		assert.NoError(t, err)

		rawRecorder := recorder.(*httpRecorder)
		assert.Equal(t, 202, w.Code)
		assert.Equal(t, 202, rawRecorder.statusCode)
		assert.Equal(t, "bar", w.Header().Get("foo"))
		assert.Equal(t, "bar", rawRecorder.Header().Get("foo"))
		assert.Equal(t, "the body", w.Body.String())
	})

	t.Run("body logger", func(t *testing.T) {
		t.Run("log error body, save body when 4xx", func(t *testing.T) {
			w := httptest.NewRecorder()
			recorder := newHTTPRecorder(w, LogErrorBody)

			recorder.WriteHeader(http.StatusBadRequest)
			_, err := recorder.Write([]byte(`{"error": "bad"}`))
			assert.NoError(t, err)

			rawRecorder := recorder.(*httpRecorder)
			assert.Equal(t, 400, w.Code)
			assert.Equal(t, 400, rawRecorder.statusCode)
			assert.Equal(t, `{"error": "bad"}`, w.Body.String())

			assert.NotNil(t, recorder.ResponseBody())
			assert.Equal(t, `{"error": "bad"}`, *recorder.ResponseBody())
		})

		t.Run("log error body, don't save body when 2xx", func(t *testing.T) {
			w := httptest.NewRecorder()
			recorder := newHTTPRecorder(w, LogErrorBody)

			recorder.WriteHeader(http.StatusOK)
			_, err := recorder.Write([]byte(`{"status": "ok"}`))
			assert.NoError(t, err)

			rawRecorder := recorder.(*httpRecorder)
			assert.Equal(t, 200, w.Code)
			assert.Equal(t, 200, rawRecorder.statusCode)
			assert.Equal(t, `{"status": "ok"}`, w.Body.String())

			assert.Nil(t, recorder.ResponseBody())
		})
	})

	t.Run("hijackable response writer", func(t *testing.T) {
		w := &dummyHijackableResponseWriter{}
		recorder := newHTTPRecorder(w, nil)

		assert.IsType(t, &hijackableRecorder{}, recorder)
		_, ok := recorder.(http.ResponseWriter)
		assert.True(t, ok, "recorder must implement http.ResponseWriter")
		_, ok = recorder.(http.Hijacker)
		assert.True(t, ok, "recorder must implement http.Hijacker")
	})
}

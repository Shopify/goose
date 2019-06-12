package srvutil

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/tomb.v2"

	"github.com/Shopify/goose/safely"
	"github.com/Shopify/goose/syncio"
)

func ExampleNewServer() {
	tb := &tomb.Tomb{}
	sl := FuncServlet("/hello/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]
		fmt.Fprintf(w, "hello %s", name)
	})

	sl = UseServlet(sl,
		// Should be first to properly add tags and logging fields to the context
		RequestContextMiddleware,
		RequestMetricsMiddleware(LogErrorBody),
		safely.Middleware,
	)

	s := NewServer(tb, "127.0.0.1:0", sl)
	defer s.Tomb().Kill(nil)
	safely.Run(s)

	u := "http://" + s.Addr().String() + "/hello/world"

	res, _ := http.Get(u)
	io.Copy(os.Stdout, res.Body)

	// Output:
	// hello world
}

func TestNewServer(t *testing.T) {
	logOutput := &syncio.Buffer{}
	origOut := logrus.StandardLogger().Out
	logrus.StandardLogger().SetOutput(logOutput)
	logrus.StandardLogger().SetLevel(logrus.DebugLevel)
	defer logrus.StandardLogger().SetOutput(origOut)

	tb := &tomb.Tomb{}
	sl := FuncServlet("/", func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		_, err := res.Write([]byte("great success"))
		assert.NoError(t, err)
	})
	s := NewServer(tb, "127.0.0.1:0", sl)
	defer s.Tomb().Kill(nil)
	safely.Run(s)

	u := "http://" + s.Addr().String()
	t.Logf("test server running on %s", u)

	assert.Contains(t, logOutput.String(), "level=info msg=\"starting server\" bind=\"127.0.0.1:0\"")

	// Works
	res, err := http.Get(u)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "great success", string(body))

	assert.NotContains(t, logOutput.String(), fmt.Sprintf("level=debug msg=\"stopped server\" addr=\"127.0.0.1:%d\" bind=\"127.0.0.1:0\"", s.Addr().Port))

	tb.Kill(errors.New("testing"))
	<-tb.Dead()

	assert.Contains(t, logOutput.String(), fmt.Sprintf("level=info msg=\"started server\" addr=\"127.0.0.1:%d\" bind=\"127.0.0.1:0\"", s.Addr().Port))
	assert.Contains(t, logOutput.String(), fmt.Sprintf("level=debug msg=\"stopped server\" addr=\"127.0.0.1:%d\" bind=\"127.0.0.1:0\"", s.Addr().Port))

	// No longer works
	res, err = http.Get(u)
	errMsg := fmt.Sprintf("Get %s: dial tcp %s: connect: connection refused", u, s.Addr().String())
	assert.EqualError(t, err, errMsg)
	assert.Nil(t, res)
}

func TestNewServerFromFactory(t *testing.T) {
	logOutput := &syncio.Buffer{}
	origOut := logrus.StandardLogger().Out
	logrus.StandardLogger().SetOutput(logOutput)
	logrus.StandardLogger().SetLevel(logrus.DebugLevel)
	defer logrus.StandardLogger().SetOutput(origOut)

	totalCallCount := 0
	servletCallCount := 0

	tb := &tomb.Tomb{}
	sl := FuncServlet("/", func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		_, err := res.Write([]byte("great success"))
		assert.NoError(t, err)
	})
	sl = UseServlet(sl, func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			servletCallCount++
			handler.ServeHTTP(w, r)
		})
	})

	s := NewServerFromFactory(tb, sl, func(handler http.Handler) http.Server {
		return http.Server{
			Addr: "127.0.0.1:0",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				totalCallCount++
				handler.ServeHTTP(w, r)
			}),
		}
	})
	defer s.Tomb().Kill(nil)
	safely.Run(s)

	u := "http://" + s.Addr().String()
	t.Logf("test server running on %s", u)

	assert.Contains(t, logOutput.String(), "level=info msg=\"starting server\" bind=\"127.0.0.1:0\"")

	// Works
	res, err := http.Get(u)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "great success", string(body))

	assert.Equal(t, 1, totalCallCount)
	assert.Equal(t, 1, servletCallCount)

	// Do some requests.
	_, err = http.Get(u)
	assert.NoError(t, err)
	_, err = http.Get(u)
	assert.NoError(t, err)
	assert.Equal(t, 3, totalCallCount)
	assert.Equal(t, 3, servletCallCount)

	// This request will only be handled by the server "total call count" middleware, not by the one specified for the servlet.
	_, err = http.Get(u + "/foobarbaz")
	assert.NoError(t, err)
	assert.Equal(t, 4, totalCallCount)
	assert.Equal(t, 3, servletCallCount)

	assert.NotContains(t, logOutput.String(), fmt.Sprintf("level=debug msg=\"stopped server\" addr=\"127.0.0.1:%d\" bind=\"127.0.0.1:0\"", s.Addr().Port))

	tb.Kill(errors.New("testing"))
	<-tb.Dead()

	assert.Contains(t, logOutput.String(), fmt.Sprintf("level=info msg=\"started server\" addr=\"127.0.0.1:%d\" bind=\"127.0.0.1:0\"", s.Addr().Port))
	assert.Contains(t, logOutput.String(), fmt.Sprintf("level=debug msg=\"stopped server\" addr=\"127.0.0.1:%d\" bind=\"127.0.0.1:0\"", s.Addr().Port))

	// No longer works
	res, err = http.Get(u)
	errMsg := fmt.Sprintf("Get %s: dial tcp %s: connect: connection refused", u, s.Addr().String())
	assert.EqualError(t, err, errMsg)
	assert.Nil(t, res)
}

func TestStoppableKeepaliveListener_Accept(t *testing.T) {
	logOutput := &syncio.Buffer{}
	origOut := logrus.StandardLogger().Out
	logrus.StandardLogger().SetOutput(logOutput)
	logrus.StandardLogger().SetLevel(logrus.DebugLevel)
	defer logrus.StandardLogger().SetOutput(origOut)

	handling := make(chan struct{})

	tb := &tomb.Tomb{}
	sl := FuncServlet("/", func(res http.ResponseWriter, req *http.Request) {
		// Notify that we are handling this request
		close(handling)

		// Wait for the server to be ask to shutdown
		<-tb.Dying()

		res.WriteHeader(http.StatusOK)
		_, err := res.Write([]byte("great success"))
		assert.NoError(t, err)
	})
	s := NewServer(tb, "127.0.0.1:0", sl)
	safely.Run(s)

	u := "http://" + s.Addr().String()
	t.Logf("test server running on %s", u)

	done := make(chan struct{})

	go func() {
		res, err := http.Get(u) // This will block on tb.Dying()
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		body, err := ioutil.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, "great success", string(body))
		close(done)
	}()

	<-handling
	assert.NotContains(t, logOutput.String(), fmt.Sprintf("level=debug msg=\"stopped server\" addr=\"127.0.0.1:%d\" bind=\"127.0.0.1:0\"", s.Addr().Port))

	tb.Kill(errors.New("testing"))

	<-done
	assert.Contains(t, logOutput.String(), fmt.Sprintf("level=info msg=\"shutting down server\" addr=\"127.0.0.1:%d\" bind=\"127.0.0.1:0\"", s.Addr().Port))

	<-tb.Dead()
	assert.Contains(t, logOutput.String(), fmt.Sprintf("level=debug msg=\"stopped server\" addr=\"127.0.0.1:%d\" bind=\"127.0.0.1:0\"", s.Addr().Port))

	// No longer works
	res, err := http.Get(u)
	errMsg := fmt.Sprintf("Get %s: dial tcp %s: connect: connection refused", u, s.Addr().String())
	assert.EqualError(t, err, errMsg)
	assert.Nil(t, res)
}

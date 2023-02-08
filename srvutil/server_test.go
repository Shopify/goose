package srvutil

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"gopkg.in/tomb.v2"

	"github.com/Shopify/goose/v2/safely"
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
		NewRequestMetricsMiddleware(&RequestMetricsMiddlewareConfig{BodyLogPredicate: LogErrorBody}),
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

	// Works
	res, err := http.Get(u)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "great success", string(body))

	tb.Kill(errors.New("testing"))
	<-tb.Dead()

	// No longer works
	res, err = http.Get(u)
	assert.NotNil(t, err)
	assert.True(t, strings.HasSuffix(err.Error(), ": connection refused"))
	assert.Nil(t, res)
}

func TestNewServerFromFactory(t *testing.T) {
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

	// Works
	res, err := http.Get(u)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	body, err := io.ReadAll(res.Body)
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

	tb.Kill(errors.New("testing"))
	<-tb.Dead()

	// No longer works
	res, err = http.Get(u)
	assert.NotNil(t, err)
	assert.True(t, strings.HasSuffix(err.Error(), ": connection refused"))
	assert.Nil(t, res)
}

func TestStoppableKeepaliveListener_Accept(t *testing.T) {
	handling := make(chan struct{})

	tb := &tomb.Tomb{}
	sl := FuncServlet("/", func(res http.ResponseWriter, req *http.Request) {
		// Notify that we are handling this request
		close(handling)

		// Wait for the server to be ask to shutdown
		<-tb.Dying()

		// Allow time for the server to be trying to exit
		time.Sleep(500 * time.Millisecond)

		select {
		case <-tb.Dead():
			t.Fatal("should not be dead already")
		default:
		}

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
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, "great success", string(body))
		close(done)
	}()

	<-handling

	tb.Kill(errors.New("testing"))

	<-done

	<-tb.Dead()

	// No longer works
	res, err := http.Get(u)
	assert.NotNil(t, err)
	assert.True(t, strings.HasSuffix(err.Error(), ": connection refused"))
	assert.Nil(t, res)
}

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

	"github.com/Shopify/goose/logger"
	"github.com/Shopify/goose/safely"
	"github.com/Shopify/goose/syncio"
)

func ExampleNewServer() {
	tb := &tomb.Tomb{}
	sl := FuncServlet("/hello/{name}", func(res http.ResponseWriter, req *http.Request) {
		name := mux.Vars(req)["name"]
		fmt.Fprintf(res, "hello %s", name)
	})

	sl = UseServlet(sl,
		// Should be first to properly add tags and logging fields to the context
		RequestContextMiddleware,
		RequestMetricsMiddleware,
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
	testLog, logOutput := buildLogger()
	origLog := log
	log = testLog
	defer func() { log = origLog }()

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
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	body, err = ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "great success", string(body))
}

func buildLogger() (logger.Logger, *syncio.Buffer) {
	buf := syncio.NewBuffer(nil)
	logrusLogger := logrus.New()
	logrusLogger.Level = logrus.DebugLevel
	logrusLogger.Out = buf
	logrusLogger.Formatter = &logrus.TextFormatter{
		DisableColors:    true,
		DisableTimestamp: true,
	}
	entry := logrus.NewEntry(logrusLogger)

	log := func(ctx logger.Valuer, err ...error) *logrus.Entry {
		return logger.ContextLog(ctx, nil, entry)
	}

	return log, buf
}

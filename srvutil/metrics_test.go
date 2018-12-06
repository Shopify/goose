package srvutil

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
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

	sl = UseServlet(sl, RequestContextMiddleware, RequestMetricsMiddleware)

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
	assert.Equal(t, []string{"route:/hello/@name", "route_name:world", "statusClass:2xx", "statusCode:200", "success:true"}, recordedTags)

	output := strings.ToLower(logging.String())
	assert.Contains(t, output, "foo: bar\\r\\n")
	assert.Contains(t, output, "foo: baz\\r\\n")
	assert.NotContains(t, output, "authorization: secret\\r\\n")
	assert.NotContains(t, output, "cookie: secret\\r\\n")
	assert.NotContains(t, output, "set-cookie: secret\\r\\n")
}

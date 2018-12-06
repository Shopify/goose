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
	"github.com/stretchr/testify/assert"
	"gopkg.in/tomb.v2"

	"github.com/Shopify/goose/safely"
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
	tb := &tomb.Tomb{}
	sl := FuncServlet("/", func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("great success"))
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
	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "great success", string(body))

	tb.Kill(errors.New("testing"))
	<-tb.Dead()

	// No longer works
	res, err = http.Get(u)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	body, err = ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "great success", string(body))
}

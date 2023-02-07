package srvutil

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"gopkg.in/tomb.v2"

	"github.com/Shopify/goose/v2/safely"
)

type dbConn interface {
	Query(sql string) interface{}
}

func NewStatsPage(dbConn dbConn) Servlet {
	return &statsPage{
		dbConn: dbConn,
	}
}

type statsPage struct {
	dbConn dbConn
}

func (s *statsPage) RegisterRouting(r *mux.Router) {
	r.Handle("/show-all", s)
}
func (s *statsPage) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	results := s.dbConn.Query("SELECT * FROM tables")
	fmt.Fprintf(res, "results: %v", results)
}

func ExampleHandlerServlet() {
	// Actually create a database connection
	var conn dbConn

	statsServlet := NewStatsPage(conn)

	// Apply a middleware to a whole servlet.
	var authMiddleware mux.MiddlewareFunc
	statsServlet = UseServlet(statsServlet, authMiddleware)

	// Register a whole Servlet under a path prefix.
	statsServlet = PrefixServlet(statsServlet, "/debug")

	// Combine all servlets together (if multiple)
	servlet := CombineServlets(
		// NewHomePage(conn),
		statsServlet,
	)

	t := &tomb.Tomb{}
	server := NewServer(t, "127.0.0.1:80", servlet)
	safely.Run(server)
}

type successHandler struct{}

func (h *successHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("great success"))
}

var h = &successHandler{}

func TestInlineModule(t *testing.T) {
	r := mux.NewRouter()
	s := InlineServlet(func(r *mux.Router) {
		r.Handle("/ok", h)
	})
	s.RegisterRouting(r)
	assert.HTTPBodyContains(t, r.ServeHTTP, "GET", "/ok", nil, "great success")
}

func TestFuncModule(t *testing.T) {
	r := mux.NewRouter()
	s := FuncServlet("/ok", h.ServeHTTP)
	s.RegisterRouting(r)
	assert.HTTPBodyContains(t, r.ServeHTTP, "GET", "/ok", nil, "great success")
}

func TestHandlerModule(t *testing.T) {
	r := mux.NewRouter()
	s := HandlerServlet("/ok", h)
	s.RegisterRouting(r)
	assert.HTTPBodyContains(t, r.ServeHTTP, "GET", "/ok", nil, "great success")
}

func TestPrefixModule(t *testing.T) {
	r := mux.NewRouter()
	hm := HandlerServlet("/ok", h)
	s := PrefixServlet(hm, "/foo")
	s.RegisterRouting(r)
	assert.HTTPBodyContains(t, r.ServeHTTP, "GET", "/foo/ok", nil, "great success")
}

type ctxKey int

const fooKey ctxKey = iota

func TestUseModule(t *testing.T) {
	r := mux.NewRouter()
	s := FuncServlet("/foo", func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		foo := ctx.Value(fooKey)
		res.Write([]byte(fmt.Sprintf("foo: %s", foo)))
	})

	s = UseServlet(s, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			ctx = context.WithValue(ctx, fooKey, "bar")
			req = req.WithContext(ctx)
			next.ServeHTTP(res, req)
		})
	})
	s.RegisterRouting(r)

	res := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/foo", nil)
	assert.NoError(t, err)
	r.ServeHTTP(res, req)

	assert.Equal(t, "foo: bar", res.Body.String())
}

func TestUseModuleCombine(t *testing.T) {
	s1 := HandlerServlet("/ok", h)
	s1 = UseServlet(s1, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.Header().Set("foo", "bar")
			next.ServeHTTP(res, req)
		})
	})

	s2 := HandlerServlet("/ok2", h)
	s2 = UseServlet(s2, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.Header().Set("foox", "barx")
			next.ServeHTTP(res, req)
		})
	})

	s2 = UseServlet(s2, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.Header().Set("fooz", "barz")
			next.ServeHTTP(res, req)
		})
	})
	s2 = PrefixServlet(s2, "/inner")

	s := CombineServlets(s1, s2)
	r := mux.NewRouter()
	s.RegisterRouting(r)

	{
		res := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/ok", nil)
		assert.NoError(t, err)
		r.ServeHTTP(res, req)

		assert.Equal(t, "great success", res.Body.String())
		assert.Equal(t, "bar", res.Header().Get("foo"), "s1 should have its middleware")
		assert.Equal(t, "", res.Header().Get("foox"), "s1 should not inherit s2's middlewares")
		assert.Equal(t, "", res.Header().Get("fooz"), "s1 should not inherit s2's middlewares")
	}

	{
		res := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/inner/ok2", nil)
		assert.NoError(t, err)
		r.ServeHTTP(res, req)

		assert.Equal(t, "great success", res.Body.String())
		assert.Equal(t, "", res.Header().Get("foo"), "s2 should not inherit s1's middleware")
		assert.Equal(t, "barx", res.Header().Get("foox"), "s2 should have its middleware")
		assert.Equal(t, "barz", res.Header().Get("fooz"), "s2 should have its middleware")
	}
}

func TestCombineModules(t *testing.T) {
	r := mux.NewRouter()
	m1 := InlineServlet(func(r *mux.Router) {
		r.Handle("/ok1", h)
	})
	m2 := FuncServlet("/ok2", h.ServeHTTP)
	s := CombineServlets(m1, m2)
	s.RegisterRouting(r)

	assert.HTTPBodyContains(t, r.ServeHTTP, "GET", "/ok1", nil, "great success")
	assert.HTTPBodyContains(t, r.ServeHTTP, "GET", "/ok2", nil, "great success")
}

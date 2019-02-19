package profiler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestNewServlet(t *testing.T) {
	s := NewServlet()
	rt := mux.NewRouter()
	s.RegisterRouting(rt)

	t.Run("pprof", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/debug/pprof", nil)
		rt.ServeHTTP(w, r)

		assert.Equal(t, http.StatusMovedPermanently, w.Code)
		assert.Equal(t, "/debug/pprof/", w.Header().Get("location"))
	})

	t.Run("pprof/", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/debug/pprof/", nil)
		rt.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "full goroutine stack dump")
	})

	t.Run("pprof/ui", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/debug/pprof/ui", nil)
		rt.ServeHTTP(w, r)

		assert.Equal(t, http.StatusMovedPermanently, w.Code)
		assert.Equal(t, "/debug/pprof/ui/", w.Header().Get("location"))
	})

	t.Run("pprof/ui/", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/debug/pprof/ui/", nil)
		rt.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "full goroutine stack dump")
	})

	t.Run("pprof/ui/heap", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/debug/pprof/ui/heap?gc=1", nil)
		rt.ServeHTTP(w, r)

		assert.Equal(t, http.StatusMovedPermanently, w.Code)
		assert.Equal(t, "/debug/pprof/ui/heap/?gc=1", w.Header().Get("location"))
	})

	t.Run("pprof/ui/trace/", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/debug/pprof/ui/trace/", nil)
		rt.ServeHTTP(w, r)

		assert.Equal(t, http.StatusMovedPermanently, w.Code)
		assert.Equal(t, "/debug/pprof/trace/", w.Header().Get("location"))
	})

	t.Run("pprof/ui/heap/", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/debug/pprof/ui/heap/?si=alloc_objects", nil)
		rt.ServeHTTP(w, r)

		if w.Code == http.StatusNotImplemented {
			// That's fine
			assert.Contains(t, w.Body.String(), "Could not execute dot; may need to install graphviz")
		} else {
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "viewer(")
			assert.Contains(t, w.Body.String(), "Type: alloc_objects")
		}
	})

	t.Run("pprof/ui/profile/flamegraph", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/debug/pprof/ui/profile/flamegraph?seconds=1", nil)
		rt.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "var data = {")
	})
}

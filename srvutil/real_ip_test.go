package srvutil_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/Shopify/goose/v2/srvutil"
)

func ExampleRealIPMiddleware() {
	r := mux.NewRouter()
	r.Use(srvutil.RealIPMiddleware())
}

func TestRealIPMiddleware(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "RemoteAddr: %s", r.RemoteAddr)
	})
	r.Use(srvutil.RealIPMiddleware())

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	req.Header.Set("X-Real-IP", "127.0.0.17")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "RemoteAddr: 127.0.0.17", w.Body.String())
}

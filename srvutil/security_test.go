package srvutil_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/Shopify/goose/srvutil"
)

func TestSecurityHeaderMiddleware(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Use(srvutil.SecurityHeaderMiddleware())

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "noopen", w.Header().Get("X-Download-Options"))
	assert.Equal(t, "none", w.Header().Get("X-Permitted-Cross-Domain-Policies"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-Xss-Protection"))
	assert.Equal(t, "origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	assert.Equal(t, "noopen", w.Header().Get("X-Download-Options"))
	assert.Equal(t, "block-all-mixed-content; upgrade-insecure-requests;", w.Header().Get("Content-Security-Policy"))
	assert.Equal(t, "max-age=631139040; includeSubdomains", w.Header().Get("Strict-Transport-Security"))
}

func TestSecurityHeaderMiddleware_AdjustHeaders(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Use(srvutil.SecurityHeaderMiddleware(func(headers map[string]string) {
		headers["Content-Security-Policy"] = "default-src *; script-src 'self' cdn.example.com; upgrade-insecure-requests"
		delete(headers, "X-Frame-Options")
	}))

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "noopen", w.Header().Get("X-Download-Options"))
	assert.Equal(t, "none", w.Header().Get("X-Permitted-Cross-Domain-Policies"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-Xss-Protection"))
	assert.Equal(t, "origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	assert.Equal(t, "noopen", w.Header().Get("X-Download-Options"))
	assert.Equal(t, "default-src *; script-src 'self' cdn.example.com; upgrade-insecure-requests", w.Header().Get("Content-Security-Policy"))
	assert.Equal(t, "max-age=631139040; includeSubdomains", w.Header().Get("Strict-Transport-Security"))
}

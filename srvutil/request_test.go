package srvutil

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/Shopify/goose/logger"
)

func TestBuildContext(t *testing.T) {
	r := newTestRequest("/path")
	ctx, id := BuildContext(r)
	entry := logger.ContextLog(ctx, nil, nil)

	assert.NotEmpty(t, id)
	assert.Equal(t, id, entry.Data[logger.UUIDKey])
	assert.Equal(t, "/path", entry.Data[PathKey])
	assert.Nil(t, entry.Data[RouteKey]) // No route info
}

func TestRequestContextMiddleware(t *testing.T) {
	var r *http.Request
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r = req
	})

	middleware := RequestContextMiddleware(handler)

	t.Run("basic", func(t *testing.T) {
		testReq := newTestRequest("/path")

		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, testReq)

		assert.NotEmpty(t, w.Header().Get(UUIDHeaderKey))
		assert.Equal(t, logger.GetLoggableValue(r.Context(), logger.UUIDKey), w.Header().Get(UUIDHeaderKey))

		assert.Empty(t, logger.GetLoggableValue(r.Context(), UserEmailKey)) // No email
		assert.Empty(t, logger.GetLoggableValue(r.Context(), RouteKey))     // No route info
		assert.Equal(t, "/path", logger.GetLoggableValue(r.Context(), PathKey))
	})

	t.Run("with email header", func(t *testing.T) {
		testReq := newTestRequest("/path")
		testReq.Header.Set(UserEmailHeaderKey, "foo@bar.com")

		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, testReq)

		assert.Equal(t, "foo@bar.com", logger.GetLoggableValue(r.Context(), UserEmailKey))
	})

	t.Run("with mux router", func(t *testing.T) {
		router := mux.NewRouter()
		router.Handle("/hello/{name:(?:[a-z]{1}){2,}}", handler) // Complicated regex to test the template extractor
		router.Use(RequestContextMiddleware)

		testReq := newTestRequest("/hello/world")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, testReq)

		assert.Equal(t, "/hello/@name", logger.GetLoggableValue(r.Context(), RouteKey))
		assert.Equal(t, "world", logger.GetLoggableValue(r.Context(), "route_name"))
		assert.Equal(t, "/hello/world", logger.GetLoggableValue(r.Context(), PathKey))
	})
}

func newTestRequest(path string) *http.Request {
	return httptest.NewRequest("GET", path, bytes.NewReader(nil))
}

func Test_replaceMatchableParts(t *testing.T) {
	tests := []struct {
		tpl     string
		want    string
		wantErr bool
	}{
		{"", "", false},
		{"/", "/", false},
		{"/{foo}", "/@foo", false},
		{"/{foo:[a-z]+}", "/@foo", false},
		{"/{foo:[a-z]+}/{bar:(?:[a-z]{2}){2}}", "/@foo/@bar", false},

		// These are weird, but we don't verify if a name is empty
		{"/{}", "/@", false},
		{"/{:}", "/@", false},

		// Errors
		{"/{foo", "", true},
		{"/foo}", "", true},
		{"/{foo}}", "", true},
		{"/{{foo}}", "", true},
		{"/{foo:{}", "", true},
		{"/}{", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.tpl, func(t *testing.T) {
			got, err := replaceMatchableParts(tt.tpl)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

package concurrency

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestThrottlerMiddleware(t *testing.T) {
	th := NewMockThrottler(true)
	m := ThrottlerMiddleware(th)

	t.Run("handle", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("ok"))
		})

		th.On("Run", mock.Anything, mock.Anything).Return(nil).Once()
		m(next).ServeHTTP(w, r)
		require.Equal(t, "ok", w.Body.String())

		th.AssertExpectations(t)
	})

	t.Run("busy", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		th.On("Run", mock.Anything, mock.Anything).Return(&ErrThrottled{WaitTime: 3 * time.Second}).Once()
		m(nil).ServeHTTP(w, r)
		require.Equal(t, "3", w.Header().Get("Retry-After"))
		require.Equal(t, http.StatusTooManyRequests, w.Code)

		th.AssertExpectations(t)
	})
}

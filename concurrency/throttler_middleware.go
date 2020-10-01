package concurrency

import (
	"errors"
	"fmt"
	"net/http"
)

func ThrottlerMiddleware(th Throttler) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := th.Run(r.Context(), func() error {
				next.ServeHTTP(w, r)
				return nil
			})

			var throttled *ErrThrottled
			if errors.As(err, &throttled) {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(throttled.WaitTime.Seconds())))
				w.WriteHeader(http.StatusTooManyRequests)
			}
		})
	}
}

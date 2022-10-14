package srvutil

import (
	"net/http"
	"time"
)

type RequestMetricsMiddlewareConfig struct {
	BodyLogPredicate BodyLogPredicateFunc
	Observer         RequestObserver
}

// NewRequestMetricsMiddleware records the time taken to serve a request, and logs request and response data.
// Example tags: statusClass:2xx, statusCode:200
// Should be added as a middleware after RequestContextMiddleware to benefit from its tags
func NewRequestMetricsMiddleware(c *RequestMetricsMiddlewareConfig) func(http.Handler) http.Handler {
	if c.Observer == nil {
		c.Observer = &DefaultRequestObserver{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Observer.BeforeRequest(r)

			recorder := newHTTPRecorder(w, c.BodyLogPredicate)

			startTime := time.Now()
			next.ServeHTTP(recorder, r)
			requestDuration := time.Since(startTime)

			c.Observer.AfterRequest(r, recorder, requestDuration)
		})
	}
}

// RequestMetricsMiddleware is here for backwards compatibility.
var RequestMetricsMiddleware = NewRequestMetricsMiddleware(&RequestMetricsMiddlewareConfig{})

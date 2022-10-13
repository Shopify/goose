package srvutil

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
	"time"
)

type BodyLogPredicateFunc func(statusCode int) bool

func LogErrorBody(statusCode int) bool {
	return statusCode >= 400
}

type HTTPRecorder interface {
	http.ResponseWriter
	StatusCode() int
	ResponseBody() *string
}

type httpRecorder struct {
	http.ResponseWriter
	statusCode int

	bodyLogPredicate BodyLogPredicateFunc
	body             bytes.Buffer
}

func (w *httpRecorder) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *httpRecorder) Write(data []byte) (int, error) {
	// If WriteHeader is never called, treat as 200, which is the underlying behaviour
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}

	if w.bodyLogPredicate != nil && w.bodyLogPredicate(w.statusCode) {
		w.body.Write(data)
	}

	return w.ResponseWriter.Write(data)
}

func (w *httpRecorder) StatusCode() int {
	return w.statusCode
}

func (w *httpRecorder) ResponseBody() *string {
	if w.body.Len() == 0 {
		return nil
	}
	s := w.body.String()
	return &s
}

type hijackableRecorder struct {
	httpRecorder
}

// Hijack implements the http.Hijacker interface, to allow for e.g. WebSockets.
func (w *hijackableRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.httpRecorder.ResponseWriter.(http.Hijacker).Hijack()
}

func newHTTPRecorder(w http.ResponseWriter, bodyLogPredicate BodyLogPredicateFunc) HTTPRecorder {
	recorder := httpRecorder{ResponseWriter: w, bodyLogPredicate: bodyLogPredicate}
	if _, ok := w.(http.Hijacker); ok {
		return &hijackableRecorder{recorder}
	}
	return &recorder
}

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

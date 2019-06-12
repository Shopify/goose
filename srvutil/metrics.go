package srvutil

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"

	"github.com/Shopify/goose/redact"

	"github.com/sirupsen/logrus"

	"github.com/Shopify/goose/metrics"
	"github.com/Shopify/goose/statsd"
)

type BodyLogPredicateFunc func(statusCode int) bool

func LogErrorBody(statusCode int) bool {
	return statusCode >= 400
}

type loggableHTTPRecorder interface {
	http.ResponseWriter
	LogFields() logrus.Fields
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

func (w *httpRecorder) LogFields() logrus.Fields {
	if w.statusCode > 0 {
		return logrus.Fields{
			"statusCode":  w.statusCode,
			"statusClass": fmt.Sprintf("%dxx", w.statusCode/100), // 2xx, 5xx, etc.
		}
	}
	return nil
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

func newHTTPRecorder(w http.ResponseWriter, bodyLogPredicate BodyLogPredicateFunc) loggableHTTPRecorder {
	recorder := httpRecorder{ResponseWriter: w, bodyLogPredicate: bodyLogPredicate}
	if _, ok := w.(http.Hijacker); ok {
		return &hijackableRecorder{recorder}
	}
	return &recorder
}

// RequestMetricsMiddleware records the time taken to serve a request, and logs request and response data.
// Example tags: statusClass:2xx, statusCode:200
// Should be added as a middleware after RequestContextMiddleware to benefit from its tags
func RequestMetricsMiddleware(bodyLogPredicate BodyLogPredicateFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			recorder := newHTTPRecorder(w, bodyLogPredicate)
			ctx = statsd.WatchingTagLoggable(ctx, recorder)
			r = r.WithContext(ctx)

			reqHeaders := redact.Headers(r.Header)

			log(ctx).
				WithField("method", r.Method).
				WithField("headers", reqHeaders).
				Info("http request")

			metrics.HTTPRequest.Time(ctx, func() error {
				next.ServeHTTP(recorder, r)
				return nil
			})

			resHeaders := redact.Headers(w.Header())

			logger := log(ctx).
				WithField("headers", resHeaders)

			if body := recorder.ResponseBody(); body != nil {
				logger = logger.WithField("responseBody", *body)
			}

			logger.Info("http response")

		})
	}

}

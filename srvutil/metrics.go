package srvutil

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/Shopify/goose/metrics"
	"github.com/Shopify/goose/statsd"
)

type httpRecorder struct {
	http.ResponseWriter
	statusCode int
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

// RequestMetricsMiddleware records the time taken to serve a request.
// Example tags: statusClass:2xx, statusCode:200
// Should be added as a middleware after RequestContextMiddleware to benefit from its tags
func RequestMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		recorder := &httpRecorder{ResponseWriter: w}
		ctx = statsd.WatchingTagLoggable(ctx, recorder)
		r = r.WithContext(ctx)

		headers := &bytes.Buffer{}
		// Ignore headers that may contains sensitive information.
		r.Header.WriteSubset(headers, map[string]bool{
			"Authorization": true,
			"Cookie":        true,
		})

		log(ctx, nil).
			WithField("method", r.Method).
			WithField("headers", headers).
			Debug("http request")

		metrics.HTTPRequest.Time(ctx, func() error {
			next.ServeHTTP(recorder, r)
			return nil
		})

		headers = &bytes.Buffer{}
		w.Header().WriteSubset(headers, map[string]bool{
			"Set-Cookie": true,
		})

		log(ctx, nil).
			WithField("headers", headers).
			Debug("http response")
	})
}

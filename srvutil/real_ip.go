package srvutil

import (
	"net/http"
)

const RealIPHeaderKey = "X-Real-IP"

func RealIPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			realIP := r.Header.Get(RealIPHeaderKey)

			if realIP != "" {
				r.RemoteAddr = realIP
			}

			next.ServeHTTP(w, r)
		})
	}
}

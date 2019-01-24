package srvutil

import (
	"net/http"
	"os"
)

// EnvVarHeaderMiddleware will expose environment variables as HTTP response headers.
// It can be used with github.com/gorilla/mux:Router.Use or wrapping a Handler.
func EnvVarHeaderMiddleware(vars map[string]string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for envVar, header := range vars {
				if val := os.Getenv(envVar); val != "" {
					w.Header().Add(header, val)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

package safely

import "net/http"

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer Recover()
		next.ServeHTTP(w, r)
	})
}

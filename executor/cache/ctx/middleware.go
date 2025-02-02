package ctx

import (
	"net/http"
)

func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(NewCtxCache(r.Context()))
		next.ServeHTTP(w, r)
	})
}

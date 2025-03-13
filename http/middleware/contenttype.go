package middleware

import (
	"net/http"
)

func WithContent(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerContentType := r.Header.Get("Content-Type")
		if headerContentType != "application/json" {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		h.ServeHTTP(w, r)
	})
}

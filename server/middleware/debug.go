package middleware

import (
	"log/slog"
	"net/http"
)

// The Debug middleware can be inserted anywhere and will print some useful
// debug information about the current request.
func Debug(printFullRequest bool, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if printFullRequest {
			slog.Debug("Debug middleware", "request", r)
		} else {
			slog.Debug("Debug middleware", "path", r.URL.Path)
		}

		h.ServeHTTP(w, r)
	})
}

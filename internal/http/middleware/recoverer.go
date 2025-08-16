package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recoverer recovers from panics and logs the error details.
// This middleware prevents the application from crashing due to unexpected panics
// and provides detailed logging for debugging purposes.
func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic details for debugging
					logger.Error("panic recovered",
						"error", err,
						"method", r.Method,
						"path", r.URL.Path,
						"remote_addr", r.RemoteAddr,
						"user_agent", r.UserAgent(),
						"stack", string(debug.Stack()),
					)

					// Return a 500 Internal Server Error to the client
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

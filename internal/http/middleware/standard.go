package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// DefaultTimeout defines the standard request timeout for the application.
// This value is used across various middleware components that require timeout configuration.
var DefaultTimeout = 60 * time.Second

// SlogLogger creates a middleware that logs HTTP requests using structured logging.
// It captures request details including method, path, status code, duration, and client information.
// The middleware also sanitises input to prevent log injection attacks.
func SlogLogger(l *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &statusWriter{ResponseWriter: w, status: 200}
			next.ServeHTTP(ww, r)

			// Sanitise path to prevent log injection attacks
			// Long paths are truncated to prevent log flooding and improve readability
			sanitisedPath := r.URL.Path
			if len(sanitisedPath) > 100 {
				sanitisedPath = sanitisedPath[:100] + "..."
			}

			// Build a request-scoped logger. Do NOT mutate the shared logger.
			reqLogger := l
			if rid, ok := RequestIDFromContext(r.Context()); ok {
				reqLogger = reqLogger.With("request_id", rid)
			}

			// Log structured request information for monitoring and debugging
			reqLogger.Info("request",
				"method", r.Method,
				"path", sanitisedPath,
				"status", ww.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"ip", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		})
	}
}

// statusWriter wraps the original ResponseWriter to capture the HTTP status code.
// This is necessary because the status code is not directly accessible from the ResponseWriter interface.
type statusWriter struct {
	http.ResponseWriter
	status int // Captures the HTTP status code for logging purposes
}

// WriteHeader captures the status code before delegating to the original ResponseWriter.
// This allows the middleware to log the actual status code returned to the client.
func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

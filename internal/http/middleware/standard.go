package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

var DefaultTimeout = 60 * time.Second

func SlogLogger(l *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &statusWriter{ResponseWriter: w, status: 200}
			next.ServeHTTP(ww, r)

			// Sanitize path to prevent log injection
			sanitizedPath := r.URL.Path
			if len(sanitizedPath) > 100 {
				sanitizedPath = sanitizedPath[:100] + "..."
			}

			l.Info("request",
				"method", r.Method,
				"path", sanitizedPath,
				"status", ww.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"ip", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) { w.status = code; w.ResponseWriter.WriteHeader(code) }

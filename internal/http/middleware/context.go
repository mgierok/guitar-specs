package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// ContextLogger provides enhanced context logging and debugging capabilities.
// This middleware logs request context information including deadlines, timeouts, and cancellation.
func ContextLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Log initial context information
			logContextInfo(logger, r, "request_started")

			// Create a response writer that captures status code
			responseWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Execute the request
			next.ServeHTTP(responseWriter, r)

			// Log completion with context details
			duration := time.Since(start)
			logContextInfo(logger, r, "request_completed",
				"status", responseWriter.statusCode,
				"duration_ms", duration.Milliseconds())
		})
	}
}

// ContextLoggerWithFields provides enhanced context logging with custom fields.
// This allows adding application-specific context information to logs.
func ContextLoggerWithFields(logger *slog.Logger, fields map[string]interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Log initial context information with custom fields
			logContextInfoWithFields(logger, r, "request_started", fields)

			// Create a response writer that captures status code
			responseWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Execute the request
			next.ServeHTTP(responseWriter, r)

			// Log completion with context details and custom fields
			duration := time.Since(start)
			completionFields := map[string]interface{}{
				"status":      responseWriter.statusCode,
				"duration_ms": duration.Milliseconds(),
			}
			// Merge custom fields with completion fields
			for k, v := range fields {
				completionFields[k] = v
			}
			logContextInfoWithFields(logger, r, "request_completed", completionFields)
		})
	}
}

// logContextInfo logs context information for a request.
func logContextInfo(logger *slog.Logger, r *http.Request, event string, args ...interface{}) {
	ctx := r.Context()

	// Extract context information
	contextInfo := extractContextInfo(ctx)

	// Prepare log arguments
	logArgs := []interface{}{
		"event", event,
		"method", r.Method,
		"path", r.URL.Path,
		"remote_addr", r.RemoteAddr,
		"user_agent", r.UserAgent(),
	}

	// Add context information
	for k, v := range contextInfo {
		logArgs = append(logArgs, k, v)
	}

	// Add additional arguments
	logArgs = append(logArgs, args...)

	logger.Info("context_log", logArgs...)
}

// logContextInfoWithFields logs context information with custom fields.
func logContextInfoWithFields(logger *slog.Logger, r *http.Request, event string, fields map[string]interface{}) {
	ctx := r.Context()

	// Extract context information
	contextInfo := extractContextInfo(ctx)

	// Prepare log arguments
	logArgs := []interface{}{
		"event", event,
		"method", r.Method,
		"path", r.URL.Path,
		"remote_addr", r.RemoteAddr,
		"user_agent", r.UserAgent(),
	}

	// Add context information
	for k, v := range contextInfo {
		logArgs = append(logArgs, k, v)
	}

	// Add custom fields
	for k, v := range fields {
		logArgs = append(logArgs, k, v)
	}

	logger.Info("context_log", logArgs...)
}

// extractContextInfo extracts useful information from the request context.
func extractContextInfo(ctx context.Context) map[string]interface{} {
	info := make(map[string]interface{})

	// Check if context has deadline
	if deadline, ok := ctx.Deadline(); ok {
		info["deadline"] = deadline.Format(time.RFC3339)
		info["deadline_from_now"] = time.Until(deadline).String()
	}

	// Check if context is done
	select {
	case <-ctx.Done():
		info["context_done"] = true
		info["context_error"] = ctx.Err().Error()
	default:
		info["context_done"] = false
	}

	// Add context type information
	switch {
	case ctx.Value("timeout") != nil:
		info["context_type"] = "timeout"
	case ctx.Value("deadline") != nil:
		info["context_type"] = "deadline"
	case ctx.Value("cancel") != nil:
		info["context_type"] = "cancellable"
	default:
		info["context_type"] = "standard"
	}

	return info
}

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

// ContextWithTimeoutCause creates a context with timeout and cause for better error handling.
// This is a convenience function that wraps context.WithTimeoutCause.
func ContextWithTimeoutCause(parent context.Context, timeout time.Duration, cause error) (context.Context, context.CancelFunc) {
	return context.WithTimeoutCause(parent, timeout, cause)
}

// ContextWithDeadlineCause creates a context with deadline and cause for better error handling.
// This is a convenience function that wraps context.WithDeadlineCause.
func ContextWithDeadlineCause(parent context.Context, deadline time.Time, cause error) (context.Context, context.CancelFunc) {
	return context.WithDeadlineCause(parent, deadline, cause)
}

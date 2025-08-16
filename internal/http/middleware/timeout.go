package middleware

import (
	"context"
	"net/http"
	"time"
)

// Timeout adds a timeout to HTTP requests.
// This middleware creates a context with timeout and cancels the request
// if it exceeds the specified duration.
func Timeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Create a new request with the timeout context
			r = r.WithContext(ctx)

			// Create a channel to signal when the request is done
			done := make(chan struct{})
			go func() {
				next.ServeHTTP(w, r)
				close(done)
			}()

			// Wait for either the request to complete or timeout to occur
			select {
			case <-done:
				// Request completed successfully
			case <-ctx.Done():
				// Timeout occurred
				http.Error(w, "Request Timeout", http.StatusRequestTimeout)
			}
		})
	}
}

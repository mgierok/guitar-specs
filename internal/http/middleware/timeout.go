package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Timeout adds a timeout to HTTP requests.
// This middleware ensures that requests don't hang indefinitely
// and provides better error context when timeouts occur.
func Timeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create context with timeout and cause
			ctx, cancel := context.WithTimeoutCause(r.Context(), timeout,
				fmt.Errorf("request timeout after %v", timeout))
			defer cancel()

			// Update request with new context
			r = r.WithContext(ctx)

			// Create channel to track request completion
			done := make(chan struct{})

			// Execute request in goroutine
			go func() {
				next.ServeHTTP(w, r)
				close(done)
			}()

			// Wait for either completion or timeout
			select {
			case <-done:
				// Request completed successfully
				return
			case <-ctx.Done():
				// Request timed out
				http.Error(w, "Request Timeout", http.StatusRequestTimeout)
				return
			}
		})
	}
}

// TimeoutWithCause adds a timeout to HTTP requests with a custom cause.
// This provides better error context for debugging and monitoring.
func TimeoutWithCause(timeout time.Duration, cause error) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create context with custom timeout cause
			ctx, cancel := context.WithTimeoutCause(r.Context(), timeout, cause)
			defer cancel()

			// Update request with new context
			r = r.WithContext(ctx)

			// Create channel to track request completion
			done := make(chan struct{})

			// Execute request in goroutine
			go func() {
				next.ServeHTTP(w, r)
				close(done)
			}()

			// Wait for either completion or timeout
			select {
			case <-done:
				// Request completed successfully
				return
			case <-ctx.Done():
				// Request timed out with custom cause
				http.Error(w, "Request Timeout", http.StatusRequestTimeout)
				return
			}
		})
	}
}

// TimeoutWithDeadline adds a timeout to HTTP requests with an absolute deadline.
// This is useful when you need to enforce a specific end time.
func TimeoutWithDeadline(deadline time.Time) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create context with absolute deadline
			ctx, cancel := context.WithDeadline(r.Context(), deadline)
			defer cancel()

			// Update request with new context
			r = r.WithContext(ctx)

			// Create channel to track request completion
			done := make(chan struct{})

			// Execute request in goroutine
			go func() {
				next.ServeHTTP(w, r)
				close(done)
			}()

			// Wait for either completion or timeout
			select {
			case <-done:
				// Request completed successfully
				return
			case <-ctx.Done():
				// Request timed out
				http.Error(w, "Request Timeout", http.StatusRequestTimeout)
				return
			}
		})
	}
}

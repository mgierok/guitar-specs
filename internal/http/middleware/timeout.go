package middleware

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
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

			// Capture downstream response to avoid writes after timeout
			crw := newCapturingResponseWriter(w)
			done := make(chan struct{})

			go func() {
				next.ServeHTTP(crw, r)
				close(done)
			}()

			// Prefer timeout when both happen nearly simultaneously
			select {
			case <-ctx.Done():
				http.Error(w, "Request Timeout", http.StatusRequestTimeout)
				return
			case <-done:
				crw.flush()
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

			crw := newCapturingResponseWriter(w)
			done := make(chan struct{})

			go func() {
				next.ServeHTTP(crw, r)
				close(done)
			}()

			select {
			case <-ctx.Done():
				http.Error(w, "Request Timeout", http.StatusRequestTimeout)
				return
			case <-done:
				crw.flush()
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

			crw := newCapturingResponseWriter(w)
			done := make(chan struct{})

			go func() {
				next.ServeHTTP(crw, r)
				close(done)
			}()

			select {
			case <-ctx.Done():
				http.Error(w, "Request Timeout", http.StatusRequestTimeout)
				return
			case <-done:
				crw.flush()
				return
			}
		})
	}
}

// capturingResponseWriter buffers downstream writes until we decide to emit.
type capturingResponseWriter struct {
	dst         http.ResponseWriter
	header      http.Header
	statusCode  int
	wroteHeader bool
	buf         bytes.Buffer
	mu          sync.Mutex
}

func newCapturingResponseWriter(w http.ResponseWriter) *capturingResponseWriter {
	return &capturingResponseWriter{
		dst:    w,
		header: make(http.Header),
	}
}

func (c *capturingResponseWriter) Header() http.Header { return c.header }

func (c *capturingResponseWriter) WriteHeader(code int) {
	if c.wroteHeader {
		return
	}
	c.wroteHeader = true
	c.statusCode = code
}

func (c *capturingResponseWriter) Write(b []byte) (int, error) {
	if !c.wroteHeader {
		c.WriteHeader(http.StatusOK)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buf.Write(b)
}

func (c *capturingResponseWriter) flush() {
	// Copy headers
	for k, vs := range c.header {
		for _, v := range vs {
			c.dst.Header().Add(k, v)
		}
	}
	if c.statusCode == 0 {
		c.statusCode = http.StatusOK
	}
	c.dst.WriteHeader(c.statusCode)
	if c.buf.Len() > 0 {
		_, _ = c.dst.Write(c.buf.Bytes())
	}
}

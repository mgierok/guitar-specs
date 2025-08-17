package middleware

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestContextLogger(t *testing.T) {
	// Create a logger that captures output
	var logOutput bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{}))

	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	t.Run("logs request start and completion", func(t *testing.T) {
		logOutput.Reset()
		middleware := ContextLogger(logger)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("User-Agent", "test-agent")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Check that both start and completion were logged
		logContent := logOutput.String()
		if !strings.Contains(logContent, "request_started") {
			t.Error("Expected request_started to be logged")
		}
		if !strings.Contains(logContent, "request_completed") {
			t.Error("Expected request_completed to be logged")
		}

		// Check that context information was logged
		if !strings.Contains(logContent, "context_type") {
			t.Error("Expected context_type to be logged")
		}
		if !strings.Contains(logContent, "context_done") {
			t.Error("Expected context_done to be logged")
		}

		// Check that request details were logged
		if !strings.Contains(logContent, "GET") {
			t.Error("Expected HTTP method to be logged")
		}
		if !strings.Contains(logContent, "/test") {
			t.Error("Expected request path to be logged")
		}
		if !strings.Contains(logContent, "127.0.0.1:12345") {
			t.Error("Expected remote address to be logged")
		}
		if !strings.Contains(logContent, "test-agent") {
			t.Error("Expected user agent to be logged")
		}

		// Check that status and duration were logged
		if !strings.Contains(logContent, "status") {
			t.Error("Expected status to be logged")
		}
		if !strings.Contains(logContent, "duration_ms") {
			t.Error("Expected duration to be logged")
		}
	})

	t.Run("logs context with timeout", func(t *testing.T) {
		logOutput.Reset()
		middleware := ContextLogger(logger)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Check that timeout context was detected
		logContent := logOutput.String()
		if !strings.Contains(logContent, "deadline") {
			t.Error("Expected deadline to be logged for timeout context")
		}
		if !strings.Contains(logContent, "deadline_from_now") {
			t.Error("Expected deadline_from_now to be logged for timeout context")
		}
	})

	t.Run("logs context with deadline", func(t *testing.T) {
		logOutput.Reset()
		middleware := ContextLogger(logger)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"

		// Create context with deadline
		deadline := time.Now().Add(100 * time.Millisecond)
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		defer cancel()
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Check that deadline context was detected
		logContent := logOutput.String()
		if !strings.Contains(logContent, "deadline") {
			t.Error("Expected deadline to be logged for deadline context")
		}
		if !strings.Contains(logContent, "deadline_from_now") {
			t.Error("Expected deadline_from_now to be logged for deadline context")
		}
	})

	t.Run("logs context cancellation", func(t *testing.T) {
		logOutput.Reset()
		middleware := ContextLogger(logger)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"

		// Create context that gets cancelled
		ctx, cancel := context.WithCancel(context.Background())
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		// Cancel context before request
		cancel()

		middleware.ServeHTTP(w, req)

		// Check that cancelled context was detected
		logContent := logOutput.String()
		if !strings.Contains(logContent, "context_done") {
			t.Error("Expected context_done to be logged")
		}
		if !strings.Contains(logContent, "context_error") {
			t.Error("Expected context_error to be logged")
		}
	})

	t.Run("handles different HTTP methods", func(t *testing.T) {
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				logOutput.Reset()
				middleware := ContextLogger(logger)(handler)

				req := httptest.NewRequest(method, "/test", nil)
				req.RemoteAddr = "127.0.0.1:12345"
				w := httptest.NewRecorder()

				middleware.ServeHTTP(w, req)

				// Check that method was logged
				logContent := logOutput.String()
				if !strings.Contains(logContent, method) {
					t.Errorf("Expected HTTP method '%s' to be logged", method)
				}
			})
		}
	})

	t.Run("captures response status codes", func(t *testing.T) {
		statusCodes := []int{200, 201, 400, 404, 500}

		for _, statusCode := range statusCodes {
			t.Run(fmt.Sprintf("status_%d", statusCode), func(t *testing.T) {
				logOutput.Reset()

				// Create handler that returns specific status
				statusHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(statusCode)
					w.Write([]byte("response"))
				})

				middleware := ContextLogger(logger)(statusHandler)

				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = "127.0.0.1:12345"
				w := httptest.NewRecorder()

				middleware.ServeHTTP(w, req)

				// Check that status was logged
				logContent := logOutput.String()
				if !strings.Contains(logContent, fmt.Sprintf("status=%d", statusCode)) {
					t.Errorf("Expected status=%d to be logged", statusCode)
				}
			})
		}
	})
}

func TestContextLoggerWithFields(t *testing.T) {
	// Create a logger that captures output
	var logOutput bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{}))

	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	t.Run("logs custom fields", func(t *testing.T) {
		logOutput.Reset()

		customFields := map[string]interface{}{
			"user_id":    "12345",
			"session_id": "abc123",
			"version":    "1.0.0",
		}

		middleware := ContextLoggerWithFields(logger, customFields)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Check that custom fields were logged
		logContent := logOutput.String()
		if !strings.Contains(logContent, "user_id") {
			t.Error("Expected user_id to be logged")
		}
		if !strings.Contains(logContent, "12345") {
			t.Error("Expected user_id value to be logged")
		}
		if !strings.Contains(logContent, "session_id") {
			t.Error("Expected session_id to be logged")
		}
		if !strings.Contains(logContent, "abc123") {
			t.Error("Expected session_id value to be logged")
		}
		if !strings.Contains(logContent, "version") {
			t.Error("Expected version to be logged")
		}
		if !strings.Contains(logContent, "1.0.0") {
			t.Error("Expected version value to be logged")
		}
	})

	t.Run("merges custom fields with completion fields", func(t *testing.T) {
		logOutput.Reset()

		customFields := map[string]interface{}{
			"request_type": "api",
			"priority":     "high",
		}

		middleware := ContextLoggerWithFields(logger, customFields)(handler)

		req := httptest.NewRequest("POST", "/api/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Check that both custom and completion fields were logged
		logContent := logOutput.String()

		// Custom fields
		if !strings.Contains(logContent, "request_type") {
			t.Error("Expected request_type to be logged")
		}
		if !strings.Contains(logContent, "api") {
			t.Error("Expected request_type value to be logged")
		}
		if !strings.Contains(logContent, "priority") {
			t.Error("Expected priority to be logged")
		}
		if !strings.Contains(logContent, "high") {
			t.Error("Expected priority value to be logged")
		}

		// Completion fields
		if !strings.Contains(logContent, "status") {
			t.Error("Expected status to be logged")
		}
		if !strings.Contains(logContent, "duration_ms") {
			t.Error("Expected duration to be logged")
		}
	})
}

func TestContextWithTimeoutCause(t *testing.T) {
	t.Run("creates context with timeout and cause", func(t *testing.T) {
		parent := context.Background()
		timeout := 50 * time.Millisecond // Shorter timeout
		cause := fmt.Errorf("test timeout")

		ctx, cancel := ContextWithTimeoutCause(parent, timeout, cause)
		defer cancel()

		// Check that context has deadline
		deadline, ok := ctx.Deadline()
		if !ok {
			t.Error("Expected context to have deadline")
		}

		// Check that deadline is approximately correct
		expectedDeadline := time.Now().Add(timeout)
		if deadline.Sub(expectedDeadline) > 10*time.Millisecond {
			t.Errorf("Expected deadline to be approximately %v, got %v", expectedDeadline, deadline)
		}

		// Wait for context to be done with timeout
		select {
		case <-ctx.Done():
			// Context is done as expected
		case <-time.After(100 * time.Millisecond):
			t.Error("Expected context to be done after timeout")
		}
	})
}

func TestContextWithDeadlineCause(t *testing.T) {
	t.Run("creates context with deadline and cause", func(t *testing.T) {
		parent := context.Background()
		deadline := time.Now().Add(50 * time.Millisecond) // Shorter deadline
		cause := fmt.Errorf("test deadline")

		ctx, cancel := ContextWithDeadlineCause(parent, deadline, cause)
		defer cancel()

		// Check that context has deadline
		actualDeadline, ok := ctx.Deadline()
		if !ok {
			t.Error("Expected context to have deadline")
		}

		// Check that deadline is correct
		if !actualDeadline.Equal(deadline) {
			t.Errorf("Expected deadline to be %v, got %v", deadline, actualDeadline)
		}

		// Wait for context to be done with timeout
		select {
		case <-ctx.Done():
			// Context is done as expected
		case <-time.After(100 * time.Millisecond):
			t.Error("Expected context to be done after deadline")
		}
	})
}

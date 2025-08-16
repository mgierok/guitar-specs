package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecoverer(t *testing.T) {
	// Create a logger that captures output
	var logOutput bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{}))

	// Create a handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Create a handler that panics with custom error
	customPanicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("custom error message")
	})

	// Create a normal handler
	normalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	t.Run("recovers from panic and returns 500", func(t *testing.T) {
		logOutput.Reset()
		middleware := Recoverer(logger)(panicHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("User-Agent", "test-agent")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Check that 500 status was returned
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", w.Code)
		}

		// Check that error message was returned
		expectedBody := "Internal Server Error"
		if strings.TrimSpace(w.Body.String()) != expectedBody {
			t.Errorf("Expected body '%s', got '%s'", expectedBody, strings.TrimSpace(w.Body.String()))
		}

		// Check that panic was logged
		logContent := logOutput.String()
		if !strings.Contains(logContent, "panic recovered") {
			t.Error("Expected panic recovery to be logged")
		}

		if !strings.Contains(logContent, "test panic") {
			t.Error("Expected panic message to be logged")
		}

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

		if !strings.Contains(logContent, "stack") {
			t.Error("Expected stack trace to be logged")
		}
	})

	t.Run("recovers from custom panic message", func(t *testing.T) {
		logOutput.Reset()
		middleware := Recoverer(logger)(customPanicHandler)

		req := httptest.NewRequest("POST", "/api/test", nil)
		req.RemoteAddr = "192.168.1.100:54321"
		req.Header.Set("User-Agent", "custom-agent")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Check that 500 status was returned
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", w.Code)
		}

		// Check that error message was returned
		expectedBody := "Internal Server Error"
		if strings.TrimSpace(w.Body.String()) != expectedBody {
			t.Errorf("Expected body '%s', got '%s'", expectedBody, strings.TrimSpace(w.Body.String()))
		}

		// Check that custom panic was logged
		logContent := logOutput.String()
		if !strings.Contains(logContent, "custom error message") {
			t.Error("Expected custom panic message to be logged")
		}

		if !strings.Contains(logContent, "POST") {
			t.Error("Expected HTTP method to be logged")
		}

		if !strings.Contains(logContent, "/api/test") {
			t.Error("Expected request path to be logged")
		}

		if !strings.Contains(logContent, "192.168.1.100:54321") {
			t.Error("Expected remote address to be logged")
		}
	})

	t.Run("allows normal requests to proceed", func(t *testing.T) {
		logOutput.Reset()
		middleware := Recoverer(logger)(normalHandler)

		req := httptest.NewRequest("GET", "/normal", nil)
		req.RemoteAddr = "10.0.0.1:8080"
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Check that normal response was returned
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		if w.Body.String() != "OK" {
			t.Errorf("Expected body 'OK', got '%s'", w.Body.String())
		}

		// Check that no panic recovery was logged
		logContent := logOutput.String()
		if strings.Contains(logContent, "panic recovered") {
			t.Error("Expected no panic recovery to be logged for normal request")
		}
	})

	t.Run("handles different HTTP methods", func(t *testing.T) {
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

		for _, method := range methods {
			logOutput.Reset()
			middleware := Recoverer(logger)(panicHandler)

			req := httptest.NewRequest(method, "/test", nil)
			req.RemoteAddr = "127.0.0.1:12345"
			w := httptest.NewRecorder()

			middleware.ServeHTTP(w, req)

			// Check that 500 status was returned for all methods
			if w.Code != http.StatusInternalServerError {
				t.Errorf("Expected status 500 for %s method, got %d", method, w.Code)
			}

			// Check that panic was logged for all methods
			logContent := logOutput.String()
			if !strings.Contains(logContent, method) {
				t.Errorf("Expected HTTP method '%s' to be logged", method)
			}
		}
	})

	t.Run("handles panic with nil error", func(t *testing.T) {
		logOutput.Reset()
		nilPanicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic(nil)
		})

		middleware := Recoverer(logger)(nilPanicHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Check that 500 status was returned
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", w.Code)
		}

		// Check that panic recovery was logged
		logContent := logOutput.String()
		if !strings.Contains(logContent, "panic recovered") {
			t.Error("Expected panic recovery to be logged")
		}
	})

	t.Run("handles panic with non-string error", func(t *testing.T) {
		logOutput.Reset()
		intPanicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic(42)
		})

		middleware := Recoverer(logger)(intPanicHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Check that 500 status was returned
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", w.Code)
		}

		// Check that panic recovery was logged
		logContent := logOutput.String()
		if !strings.Contains(logContent, "panic recovered") {
			t.Error("Expected panic recovery to be logged")
		}

		if !strings.Contains(logContent, "42") {
			t.Error("Expected panic value to be logged")
		}
	})
}

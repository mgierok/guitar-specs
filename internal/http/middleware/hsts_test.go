package middleware

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHSTS(t *testing.T) {
	// Create a simple handler that returns 200 OK
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with HSTS middleware
	middleware := HSTS(handler)

	t.Run("HTTPS request should include HSTS header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.TLS = &tls.ConnectionState{} // Simulate HTTPS
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		hstsHeader := w.Header().Get("Strict-Transport-Security")
		if hstsHeader == "" {
			t.Error("Expected HSTS header to be set for HTTPS request")
		}

		expectedHSTS := "max-age=31536000; includeSubDomains; preload"
		if hstsHeader != expectedHSTS {
			t.Errorf("Expected HSTS header '%s', got '%s'", expectedHSTS, hstsHeader)
		}
	})

	t.Run("HTTP request should not include HSTS header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		// No TLS field - HTTP request
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		hstsHeader := w.Header().Get("Strict-Transport-Security")
		if hstsHeader != "" {
			t.Errorf("Expected no HSTS header for HTTP request, got '%s'", hstsHeader)
		}
	})

	t.Run("HSTS header should have correct values", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/test", nil)
		req.TLS = &tls.ConnectionState{} // Simulate HTTPS
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		hstsHeader := w.Header().Get("Strict-Transport-Security")
		
		// Check that all required directives are present
		if !contains(hstsHeader, "max-age=31536000") {
			t.Error("HSTS header should contain max-age=31536000")
		}
		if !contains(hstsHeader, "includeSubDomains") {
			t.Error("HSTS header should contain includeSubDomains")
		}
		if !contains(hstsHeader, "preload") {
			t.Error("HSTS header should contain preload")
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		contains(s[1:], substr))))
}


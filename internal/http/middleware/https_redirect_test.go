package middleware

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPSRedirect(t *testing.T) {
	// Create a simple handler that returns 200 OK
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with HTTPS redirect middleware
	middleware := HTTPSRedirect(handler)

	t.Run("HTTPS request should proceed normally", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.TLS = &tls.ConnectionState{} // Simulate HTTPS
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		if w.Body.String() != "OK" {
			t.Errorf("Expected body 'OK', got '%s'", w.Body.String())
		}
	})

	t.Run("HTTP request should redirect to HTTPS", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test?param=value", nil)
		req.Host = "example.com"
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		if w.Code != http.StatusMovedPermanently {
			t.Errorf("Expected status 301, got %d", w.Code)
		}

		expectedLocation := "https://example.com/test?param=value"
		if location := w.Header().Get("Location"); location != expectedLocation {
			t.Errorf("Expected Location header '%s', got '%s'", expectedLocation, location)
		}
	})

	t.Run("HTTP request with path should redirect correctly", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/users/123", nil)
		req.Host = "localhost:8080"
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		if w.Code != http.StatusMovedPermanently {
			t.Errorf("Expected status 301, got %d", w.Code)
		}

		expectedLocation := "https://localhost:8080/api/users/123"
		if location := w.Header().Get("Location"); location != expectedLocation {
			t.Errorf("Expected Location header '%s', got '%s'", expectedLocation, location)
		}
	})
}


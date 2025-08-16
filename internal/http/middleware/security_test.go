package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeaders(t *testing.T) {
	// Create a simple handler that returns 200 OK
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with security headers middleware
	middleware := SecurityHeaders(handler)

	// Test that security headers are set correctly
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	// Check that all required security headers are set
	headers := map[string]string{
		"X-Frame-Options":         "DENY",
		"X-Content-Type-Options":  "nosniff",
		"X-XSS-Protection":        "1; mode=block",
		"Referrer-Policy":         "strict-origin-when-cross-origin",
		"Content-Security-Policy": "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; font-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'",
		"Permissions-Policy":      "geolocation=(), microphone=(), camera=()",
	}

	for header, expectedValue := range headers {
		if value := w.Header().Get(header); value != expectedValue {
			t.Errorf("Expected header %s to be '%s', got '%s'", header, expectedValue, value)
		}
	}

	// Verify response body is preserved
	if w.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got '%s'", w.Body.String())
	}
}

func TestSecurityHeadersPreservesExistingHeaders(t *testing.T) {
	// Create a handler that sets custom headers
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Custom-Header", "custom-value")
		w.Header().Set("X-Custom-Security", "custom-security")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with security headers middleware
	middleware := SecurityHeaders(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	// Check that custom headers are preserved
	if value := w.Header().Get("Custom-Header"); value != "custom-value" {
		t.Errorf("Expected Custom-Header to be 'custom-value', got '%s'", value)
	}

	if value := w.Header().Get("X-Custom-Security"); value != "custom-security" {
		t.Errorf("Expected X-Custom-Security to be 'custom-security', got '%s'", value)
	}

	// Check that security headers are still set
	if value := w.Header().Get("X-Frame-Options"); value != "DENY" {
		t.Errorf("Expected X-Frame-Options to be 'DENY', got '%s'", value)
	}
}

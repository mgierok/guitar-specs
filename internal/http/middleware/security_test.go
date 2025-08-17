package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

	// Check static security headers
	if value := w.Header().Get("X-Frame-Options"); value != "DENY" {
		t.Errorf("Expected X-Frame-Options to be 'DENY', got '%s'", value)
	}
	if value := w.Header().Get("X-Content-Type-Options"); value != "nosniff" {
		t.Errorf("Expected X-Content-Type-Options to be 'nosniff', got '%s'", value)
	}
	if value := w.Header().Get("X-XSS-Protection"); value != "1; mode=block" {
		t.Errorf("Expected X-XSS-Protection to be '1; mode=block', got '%s'", value)
	}
	if value := w.Header().Get("Referrer-Policy"); value != "strict-origin-when-cross-origin" {
		t.Errorf("Expected Referrer-Policy to be 'strict-origin-when-cross-origin', got '%s'", value)
	}
	if value := w.Header().Get("Permissions-Policy"); value != "geolocation=(), microphone=(), camera=()" {
		t.Errorf("Expected Permissions-Policy to be 'geolocation=(), microphone=(), camera=()', got '%s'", value)
	}

	// Check CSP contains nonce and core directives
	csp := w.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Fatalf("Expected Content-Security-Policy header to be set")
	}
	if !strings.Contains(csp, "default-src 'self';") {
		t.Errorf("CSP missing default-src: %s", csp)
	}
	if !strings.Contains(csp, "script-src 'self' 'nonce-") {
		t.Errorf("CSP missing script-src nonce: %s", csp)
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

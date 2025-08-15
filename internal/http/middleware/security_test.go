package middleware

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeaders(t *testing.T) {
	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	// Wrap with SecurityHeaders middleware
	middleware := SecurityHeaders(handler)

	// Create test request
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	// Execute request
	middleware.ServeHTTP(rec, req)

	// Check all required security headers
	headers := map[string]string{
		"X-Frame-Options":         "DENY",
		"X-Content-Type-Options":  "nosniff",
		"X-XSS-Protection":        "1; mode=block",
		"Referrer-Policy":         "strict-origin-when-cross-origin",
		"Content-Security-Policy": "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; font-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'",
		"Permissions-Policy":      "geolocation=(), microphone=(), camera=()",
	}

	for header, expectedValue := range headers {
		actualValue := rec.Header().Get(header)
		if actualValue != expectedValue {
			t.Errorf("Header %s: expected %s, got %s", header, expectedValue, actualValue)
		}
	}

	// Check that HSTS is NOT set for HTTP (non-TLS) requests
	hsts := rec.Header().Get("Strict-Transport-Security")
	if hsts != "" {
		t.Errorf("HSTS header should not be set for HTTP requests, got: %s", hsts)
	}
}

func TestSecurityHeadersHSTS(t *testing.T) {
	// Test that HSTS is set for HTTPS requests
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	middleware := SecurityHeaders(handler)

	// Create HTTPS request (simulate TLS)
	req := httptest.NewRequest("GET", "https://localhost/", nil)
	req.TLS = &tls.ConnectionState{} // Simulate TLS connection
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	// Check that HSTS is set for HTTPS
	hsts := rec.Header().Get("Strict-Transport-Security")
	expectedHSTS := "max-age=15552000; includeSubDomains; preload"
	if hsts != expectedHSTS {
		t.Errorf("HSTS header: expected %s, got %s", expectedHSTS, hsts)
	}
}

func TestSecurityHeadersNonGET(t *testing.T) {
	// Test that security headers are set for all HTTP methods
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	middleware := SecurityHeaders(handler)

	methods := []string{"POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		req := httptest.NewRequest(method, "/", nil)
		rec := httptest.NewRecorder()

		middleware.ServeHTTP(rec, req)

		// Check that security headers are present for all methods
		csp := rec.Header().Get("Content-Security-Policy")
		if csp == "" {
			t.Errorf("Content-Security-Policy header missing for %s request", method)
		}
	}
}

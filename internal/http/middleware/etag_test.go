package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestETag(t *testing.T) {
	// Create a simple handler that returns some content
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	// Wrap with ETag middleware
	middleware := ETag(handler)

	// Create test request
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	// Execute request
	middleware.ServeHTTP(rec, req)

	// Check if ETag header is present
	etag := rec.Header().Get("ETag")
	if etag == "" {
		t.Error("ETag header not found")
	}

	// Check if ETag has correct format (quoted hex string)
	if len(etag) < 3 || etag[0] != '"' || etag[len(etag)-1] != '"' {
		t.Errorf("ETag has wrong format: %s", etag)
	}

	t.Logf("ETag header: %s", etag)
}

func TestETagNonGET(t *testing.T) {
	// Test that non-GET requests don't get ETag headers
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	middleware := ETag(handler)

	req := httptest.NewRequest("POST", "/", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	etag := rec.Header().Get("ETag")
	if etag != "" {
		t.Errorf("ETag header should not be present for POST request, got: %s", etag)
	}
}

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	// Create rate limiter: 3 requests per 100ms
	rateLimiter := NewRateLimiter(3, 100*time.Millisecond)
	middleware := rateLimiter.RateLimit(handler)

	// Test successful requests within limit
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rec := httptest.NewRecorder()

		middleware.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i+1, rec.Code)
		}
	}

	// Test that 4th request is blocked
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("4th request: expected status 429, got %d", rec.Code)
	}
}

func TestRateLimiterDifferentIPs(t *testing.T) {
	// Test that different IPs have separate limits
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	rateLimiter := NewRateLimiter(2, 100*time.Millisecond)
	middleware := rateLimiter.RateLimit(handler)

	// IP 1: 2 requests (should succeed)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rec := httptest.NewRecorder()

		middleware.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("IP1 request %d: expected status 200, got %d", i+1, rec.Code)
		}
	}

	// IP 2: 2 requests (should succeed - different IP)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.2:12345"
		rec := httptest.NewRecorder()

		middleware.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("IP2 request %d: expected status 200, got %d", i+1, rec.Code)
		}
	}
}

func TestRateLimiterWindowExpiry(t *testing.T) {
	// Test that rate limit resets after window expires
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	rateLimiter := NewRateLimiter(2, 50*time.Millisecond)
	middleware := rateLimiter.RateLimit(handler)

	// Make 2 requests (should succeed)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rec := httptest.NewRecorder()

		middleware.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i+1, rec.Code)
		}
	}

	// 3rd request should be blocked
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("3rd request: expected status 429, got %d", rec.Code)
	}

	// Wait for window to expire
	time.Sleep(60 * time.Millisecond)

	// Next request should succeed (window expired)
	req = httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec = httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Request after expiry: expected status 200, got %d", rec.Code)
	}
}

func TestRateLimiterCleanup(t *testing.T) {
	// Test that old entries are cleaned up
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	rateLimiter := NewRateLimiter(1, 10*time.Millisecond)
	middleware := rateLimiter.RateLimit(handler)

	// Make request
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("First request: expected status 200, got %d", rec.Code)
	}

	// Wait for window to expire
	time.Sleep(15 * time.Millisecond)

	// Make another request (should succeed as old entry was cleaned up)
	req = httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec = httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Request after cleanup: expected status 200, got %d", rec.Code)
	}
}

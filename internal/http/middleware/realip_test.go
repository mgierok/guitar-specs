package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRealIP(t *testing.T) {
	// Create a simple handler that returns the remote address
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(r.RemoteAddr))
	})

	t.Run("extracts X-Forwarded-For header", func(t *testing.T) {
		trustedProxies := []string{"127.0.0.1", "::1"}
		middleware := RealIP(trustedProxies)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("X-Forwarded-For", "203.0.113.1, 127.0.0.1")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Check that the first IP from X-Forwarded-For was used
		expectedIP := "203.0.113.1"
		if req.RemoteAddr != expectedIP {
			t.Errorf("Expected RemoteAddr to be '%s', got '%s'", expectedIP, req.RemoteAddr)
		}

		if w.Body.String() != expectedIP {
			t.Errorf("Expected response body '%s', got '%s'", expectedIP, w.Body.String())
		}
	})

	t.Run("extracts X-Real-IP header", func(t *testing.T) {
		trustedProxies := []string{"127.0.0.1", "::1"}
		middleware := RealIP(trustedProxies)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("X-Real-IP", "198.51.100.1")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		expectedIP := "198.51.100.1"
		if req.RemoteAddr != expectedIP {
			t.Errorf("Expected RemoteAddr to be '%s', got '%s'", expectedIP, req.RemoteAddr)
		}

		if w.Body.String() != expectedIP {
			t.Errorf("Expected response body '%s', got '%s'", expectedIP, w.Body.String())
		}
	})

	t.Run("extracts X-Client-IP header", func(t *testing.T) {
		trustedProxies := []string{"127.0.0.1", "::1"}
		middleware := RealIP(trustedProxies)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("X-Client-IP", "192.168.1.100")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		expectedIP := "192.168.1.100"
		if req.RemoteAddr != expectedIP {
			t.Errorf("Expected RemoteAddr to be '%s', got '%s'", expectedIP, req.RemoteAddr)
		}

		if w.Body.String() != expectedIP {
			t.Errorf("Expected response body '%s', got '%s'", expectedIP, w.Body.String())
		}
	})

	t.Run("extracts CF-Connecting-IP header (Cloudflare)", func(t *testing.T) {
		trustedProxies := []string{"127.0.0.1", "::1"}
		middleware := RealIP(trustedProxies)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("CF-Connecting-IP", "104.16.123.45")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		expectedIP := "104.16.123.45"
		if req.RemoteAddr != expectedIP {
			t.Errorf("Expected RemoteAddr to be '%s', got '%s'", expectedIP, req.RemoteAddr)
		}

		if w.Body.String() != expectedIP {
			t.Errorf("Expected response body '%s', got '%s'", expectedIP, w.Body.String())
		}
	})

	t.Run("prioritizes X-Forwarded-For over other headers", func(t *testing.T) {
		trustedProxies := []string{"127.0.0.1", "::1"}
		middleware := RealIP(trustedProxies)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("X-Forwarded-For", "203.0.113.1")
		req.Header.Set("X-Real-IP", "198.51.100.1")
		req.Header.Set("X-Client-IP", "192.168.1.100")
		req.Header.Set("CF-Connecting-IP", "104.16.123.45")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// X-Forwarded-For should take priority
		expectedIP := "203.0.113.1"
		if req.RemoteAddr != expectedIP {
			t.Errorf("Expected RemoteAddr to be '%s', got '%s'", expectedIP, req.RemoteAddr)
		}
	})

	t.Run("falls back to direct connection IP when no headers", func(t *testing.T) {
		trustedProxies := []string{"127.0.0.1", "::1"}
		middleware := RealIP(trustedProxies)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "203.0.113.1:12345"
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		expectedIP := "203.0.113.1:12345"
		if req.RemoteAddr != expectedIP {
			t.Errorf("Expected RemoteAddr to be '%s', got '%s'", expectedIP, req.RemoteAddr)
		}

		if w.Body.String() != expectedIP {
			t.Errorf("Expected response body '%s', got '%s'", expectedIP, w.Body.String())
		}
	})

	t.Run("handles empty X-Forwarded-For header", func(t *testing.T) {
		trustedProxies := []string{"127.0.0.1", "::1"}
		middleware := RealIP(trustedProxies)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "203.0.113.1:12345"
		req.Header.Set("X-Forwarded-For", "")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		expectedIP := "203.0.113.1:12345"
		if req.RemoteAddr != expectedIP {
			t.Errorf("Expected RemoteAddr to be '%s', got '%s'", expectedIP, req.RemoteAddr)
		}
	})

	t.Run("handles malformed X-Forwarded-For header", func(t *testing.T) {
		trustedProxies := []string{"127.0.0.1", "::1"}
		middleware := RealIP(trustedProxies)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "203.0.113.1:12345"
		req.Header.Set("X-Forwarded-For", "invalid-ip, 127.0.0.1")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Should fall back to direct connection IP
		expectedIP := "203.0.113.1:12345"
		if req.RemoteAddr != expectedIP {
			t.Errorf("Expected RemoteAddr to be '%s', got '%s'", expectedIP, req.RemoteAddr)
		}
	})

	t.Run("handles private IP addresses in X-Forwarded-For", func(t *testing.T) {
		trustedProxies := []string{"127.0.0.1", "::1"}
		middleware := RealIP(trustedProxies)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "203.0.113.1:12345"
		req.Header.Set("X-Forwarded-For", "192.168.1.100, 127.0.0.1")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Should fall back to direct connection IP since 203.0.113.1 is not trusted
		expectedIP := "203.0.113.1:12345"
		if req.RemoteAddr != expectedIP {
			t.Errorf("Expected RemoteAddr to be '%s', got '%s'", expectedIP, req.RemoteAddr)
		}
	})

	t.Run("rejects headers from untrusted proxies", func(t *testing.T) {
		trustedProxies := []string{"127.0.0.1", "::1"}
		middleware := RealIP(trustedProxies)(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "203.0.113.1:12345" // Not in trusted list
		req.Header.Set("X-Forwarded-For", "192.168.1.100")
		req.Header.Set("X-Real-IP", "198.51.100.1")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Should use direct connection IP since proxy is not trusted
		expectedIP := "203.0.113.1:12345"
		if req.RemoteAddr != expectedIP {
			t.Errorf("Expected RemoteAddr to be '%s', got '%s'", expectedIP, req.RemoteAddr)
		}
	})
}

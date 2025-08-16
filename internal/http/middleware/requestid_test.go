package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID(t *testing.T) {
	// Create a simple handler that returns 200 OK
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with RequestID middleware
	middleware := RequestID(handler)

	t.Run("generates new request ID when none exists", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Check that request ID was generated
		requestID := req.Header.Get("X-Request-ID")
		if requestID == "" {
			t.Error("Expected request ID to be generated")
		}

		// Check that response has the same request ID
		responseID := w.Header().Get("X-Request-ID")
		if responseID != requestID {
			t.Errorf("Expected response ID '%s' to match request ID '%s'", responseID, requestID)
		}

		// Check that request ID is 16 characters (8 bytes hex)
		if len(requestID) != 16 {
			t.Errorf("Expected request ID to be 16 characters, got %d", len(requestID))
		}
	})

	t.Run("preserves existing request ID", func(t *testing.T) {
		existingID := "existing-request-id-123"
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", existingID)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Check that existing request ID was preserved
		requestID := req.Header.Get("X-Request-ID")
		if requestID != existingID {
			t.Errorf("Expected existing request ID '%s' to be preserved, got '%s'", existingID, requestID)
		}

		// Check that response has the same request ID
		responseID := w.Header().Get("X-Request-ID")
		if responseID != existingID {
			t.Errorf("Expected response ID '%s' to match existing request ID '%s'", responseID, existingID)
		}
	})

	t.Run("generates unique request IDs", func(t *testing.T) {
		ids := make(map[string]bool)
		const numRequests = 100

		for i := 0; i < numRequests; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			middleware.ServeHTTP(w, req)

			requestID := req.Header.Get("X-Request-ID")
			if ids[requestID] {
				t.Errorf("Duplicate request ID generated: %s", requestID)
			}
			ids[requestID] = true
		}

		if len(ids) != numRequests {
			t.Errorf("Expected %d unique IDs, got %d", numRequests, len(ids))
		}
	})

	t.Run("handles different HTTP methods", func(t *testing.T) {
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

		for _, method := range methods {
			req := httptest.NewRequest(method, "/test", nil)
			w := httptest.NewRecorder()

			middleware.ServeHTTP(w, req)

			requestID := req.Header.Get("X-Request-ID")
			if requestID == "" {
				t.Errorf("Expected request ID for %s method", method)
			}

			responseID := w.Header().Get("X-Request-ID")
			if responseID != requestID {
				t.Errorf("Expected response ID to match request ID for %s method", method)
			}
		}
	})

	t.Run("preserves response body", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		if w.Body.String() != "OK" {
			t.Errorf("Expected response body 'OK', got '%s'", w.Body.String())
		}

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})
}

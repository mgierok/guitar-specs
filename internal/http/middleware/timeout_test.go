package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTimeout(t *testing.T) {
	t.Run("allows request to complete within timeout", func(t *testing.T) {
		// Create a handler that completes quickly
		fastHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		middleware := Timeout(100 * time.Millisecond)(fastHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Check that request completed successfully
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		if w.Body.String() != "OK" {
			t.Errorf("Expected body 'OK', got '%s'", w.Body.String())
		}
	})

	t.Run("returns timeout error when request exceeds timeout", func(t *testing.T) {
		// Create a handler that takes longer than the timeout
		slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(200 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		middleware := Timeout(50 * time.Millisecond)(slowHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Check that timeout error was returned
		if w.Code != http.StatusRequestTimeout {
			t.Errorf("Expected status 408, got %d", w.Code)
		}

		expectedBody := "Request Timeout"
		if strings.TrimSpace(w.Body.String()) != expectedBody {
			t.Errorf("Expected body '%s', got '%s'", expectedBody, strings.TrimSpace(w.Body.String()))
		}
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		// Create a handler that waits for context cancellation
		contextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			<-r.Context().Done()
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		middleware := Timeout(10 * time.Millisecond)(contextHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Check that timeout error was returned
		if w.Code != http.StatusRequestTimeout {
			t.Errorf("Expected status 408, got %d", w.Code)
		}
	})

	t.Run("preserves request context", func(t *testing.T) {
		// Create a handler that checks the context
		contextCheckHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check that the context has a timeout
			_, ok := r.Context().Deadline()
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("No deadline in context"))
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Context OK"))
		})

		middleware := Timeout(100 * time.Millisecond)(contextCheckHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Check that request completed successfully
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		if w.Body.String() != "Context OK" {
			t.Errorf("Expected body 'Context OK', got '%s'", w.Body.String())
		}
	})

	t.Run("handles different timeout values", func(t *testing.T) {
		timeouts := []time.Duration{
			10 * time.Millisecond,
			50 * time.Millisecond,
			100 * time.Millisecond,
			500 * time.Millisecond,
		}

		for _, timeout := range timeouts {
			t.Run(timeout.String(), func(t *testing.T) {
				// Create a handler that takes exactly the timeout duration
				exactHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(timeout)
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("OK"))
				})

				// Set timeout slightly shorter to trigger timeout
				middleware := Timeout(timeout - 10*time.Millisecond)(exactHandler)

				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()

				middleware.ServeHTTP(w, req)

				// Check that timeout error was returned
				if w.Code != http.StatusRequestTimeout {
					t.Errorf("Expected status 408 for timeout %v, got %d", timeout, w.Code)
				}
			})
		}
	})

	t.Run("handles zero timeout", func(t *testing.T) {
		// Create a handler that completes quickly
		fastHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		middleware := Timeout(0)(fastHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Check that timeout error was returned immediately
		if w.Code != http.StatusRequestTimeout {
			t.Errorf("Expected status 408 for zero timeout, got %d", w.Code)
		}
	})

	t.Run("handles very short timeout", func(t *testing.T) {
		// Create a handler that takes a very short time
		veryFastHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(1 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		middleware := Timeout(5 * time.Millisecond)(veryFastHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Check that request completed successfully
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		if w.Body.String() != "OK" {
			t.Errorf("Expected body 'OK', got '%s'", w.Body.String())
		}
	})

	t.Run("handles different HTTP methods", func(t *testing.T) {
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				// Create a handler that takes longer than the timeout
				slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(200 * time.Millisecond)
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("OK"))
				})

				middleware := Timeout(50 * time.Millisecond)(slowHandler)

				req := httptest.NewRequest(method, "/test", nil)
				w := httptest.NewRecorder()

				middleware.ServeHTTP(w, req)

				// Check that timeout error was returned for all methods
				if w.Code != http.StatusRequestTimeout {
					t.Errorf("Expected status 408 for %s method, got %d", method, w.Code)
				}
			})
		}
	})
}

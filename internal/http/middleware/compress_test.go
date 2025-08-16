package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCompress(t *testing.T) {
	t.Run("compresses supported content types", func(t *testing.T) {
		// Create a handler that sets content type and returns content
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("<html><body>Hello World</body></html>"))
		})

		middleware := Compress(5, "text/html")(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Check that response was compressed
		if w.Header().Get("Content-Encoding") != "gzip" {
			t.Error("Expected Content-Encoding: gzip header")
		}

		if !strings.Contains(w.Header().Get("Vary"), "Accept-Encoding") {
			t.Error("Expected Vary: Accept-Encoding header")
		}

		// Check that content was actually compressed
		if w.Body.Len() == 0 {
			t.Error("Expected compressed content")
		}

		// Verify we can decompress the content
		reader, err := gzip.NewReader(w.Body)
		if err != nil {
			t.Fatalf("Failed to create gzip reader: %v", err)
		}
		defer reader.Close()

		decompressed, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("Failed to decompress content: %v", err)
		}

		expected := "<html><body>Hello World</body></html>"
		if string(decompressed) != expected {
			t.Errorf("Expected decompressed content '%s', got '%s'", expected, string(decompressed))
		}
	})

	t.Run("does not compress unsupported content types", func(t *testing.T) {
		// Create a handler that sets unsupported content type
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/png")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("fake png data"))
		})

		middleware := Compress(5, "text/html", "text/css")(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Check that response was not compressed
		if w.Header().Get("Content-Encoding") == "gzip" {
			t.Error("Expected no Content-Encoding: gzip header for unsupported content type")
		}

		if w.Header().Get("Vary") != "" {
			t.Error("Expected no Vary header for unsupported content type")
		}

		// Check that content was not compressed
		expected := "fake png data"
		if w.Body.String() != expected {
			t.Errorf("Expected uncompressed content '%s', got '%s'", expected, w.Body.String())
		}
	})

	t.Run("does not compress when client doesn't support gzip", func(t *testing.T) {
		// Create a handler that sets supported content type
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("<html><body>Hello World</body></html>"))
		})

		middleware := Compress(5, "text/html")(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		// No Accept-Encoding header
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Check that response was not compressed
		if w.Header().Get("Content-Encoding") == "gzip" {
			t.Error("Expected no Content-Encoding: gzip header when client doesn't support gzip")
		}

		if w.Header().Get("Vary") != "" {
			t.Error("Expected no Vary header when client doesn't support gzip")
		}

		// Check that content was not compressed
		expected := "<html><body>Hello World</body></html>"
		if w.Body.String() != expected {
			t.Errorf("Expected uncompressed content '%s', got '%s'", expected, w.Body.String())
		}
	})

	t.Run("handles content type not known upfront", func(t *testing.T) {
		// Create a handler that doesn't set content type initially
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Write some content first
			w.Write([]byte("some content"))
			// Then set content type
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
		})

		middleware := Compress(5, "text/html")(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Check that response was compressed
		if w.Header().Get("Content-Encoding") != "gzip" {
			t.Error("Expected Content-Encoding: gzip header for content type set after write")
		}

		if !strings.Contains(w.Header().Get("Vary"), "Accept-Encoding") {
			t.Error("Expected Vary: Accept-Encoding header")
		}
	})

	t.Run("handles compression failure gracefully", func(t *testing.T) {
		// Create a handler that sets supported content type
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("<html><body>Hello World</body></html>"))
		})

		// Use invalid compression level to trigger failure
		middleware := Compress(999, "text/html")(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Check that response was not compressed due to failure
		if w.Header().Get("Content-Encoding") == "gzip" {
			t.Error("Expected no Content-Encoding: gzip header when compression fails")
		}

		// Check that content was served uncompressed
		expected := "<html><body>Hello World</body></html>"
		if w.Body.String() != expected {
			t.Errorf("Expected uncompressed content '%s', got '%s'", expected, w.Body.String())
		}
	})

	t.Run("handles different compression levels", func(t *testing.T) {
		levels := []int{1, 5, 9}

		for _, level := range levels {
			t.Run("level_"+string(rune(level+'0')), func(t *testing.T) {
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "text/html")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("<html><body>Hello World</body></html>"))
				})

				middleware := Compress(level, "text/html")(handler)

				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Accept-Encoding", "gzip")
				w := httptest.NewRecorder()

				middleware.ServeHTTP(w, req)

				// Check that response was compressed
				if w.Header().Get("Content-Encoding") != "gzip" {
					t.Errorf("Expected Content-Encoding: gzip header for level %d", level)
				}

				// Verify we can decompress the content
				reader, err := gzip.NewReader(w.Body)
				if err != nil {
					t.Fatalf("Failed to create gzip reader for level %d: %v", level, err)
				}
				defer reader.Close()

				decompressed, err := io.ReadAll(reader)
				if err != nil {
					t.Fatalf("Failed to decompress content for level %d: %v", level, err)
				}

				expected := "<html><body>Hello World</body></html>"
				if string(decompressed) != expected {
					t.Errorf("Expected decompressed content '%s' for level %d, got '%s'", expected, level, string(decompressed))
				}
			})
		}
	})

	t.Run("handles prefix matching for content types", func(t *testing.T) {
		// Create a handler that sets content type with prefix
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/css")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("body { color: red; }"))
		})

		// Use prefix matching
		middleware := Compress(5, "text/")(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Check that response was compressed
		if w.Header().Get("Content-Encoding") != "gzip" {
			t.Error("Expected Content-Encoding: gzip header for prefix-matched content type")
		}

		// Verify we can decompress the content
		reader, err := gzip.NewReader(w.Body)
		if err != nil {
			t.Fatalf("Failed to create gzip reader: %v", err)
		}
		defer reader.Close()

		decompressed, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("Failed to decompress content: %v", err)
		}

		expected := "body { color: red; }"
		if string(decompressed) != expected {
			t.Errorf("Expected decompressed content '%s', got '%s'", expected, string(decompressed))
		}
	})

	t.Run("handles multiple content types", func(t *testing.T) {
		contentTypes := []string{"text/html", "text/css", "application/javascript"}

		for _, contentType := range contentTypes {
			t.Run(contentType, func(t *testing.T) {
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", contentType)
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("test content"))
				})

				middleware := Compress(5, "text/html", "text/css", "application/javascript")(handler)

				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Accept-Encoding", "gzip")
				w := httptest.NewRecorder()

				middleware.ServeHTTP(w, req)

				// Check that response was compressed
				if w.Header().Get("Content-Encoding") != "gzip" {
					t.Errorf("Expected Content-Encoding: gzip header for %s", contentType)
				}
			})
		}
	})

	t.Run("preserves response status codes", func(t *testing.T) {
		statusCodes := []int{200, 201, 400, 404, 500}

		for _, statusCode := range statusCodes {
			t.Run("status_"+string(rune(statusCode/100+'0'))+string(rune(statusCode%100/10+'0'))+string(rune(statusCode%10+'0')), func(t *testing.T) {
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "text/html")
					w.WriteHeader(statusCode)
					w.Write([]byte("error content"))
				})

				middleware := Compress(5, "text/html")(handler)

				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Accept-Encoding", "gzip")
				w := httptest.NewRecorder()

				middleware.ServeHTTP(w, req)

				// Check that status code was preserved
				if w.Code != statusCode {
					t.Errorf("Expected status code %d, got %d", statusCode, w.Code)
				}
			})
		}
	})
}

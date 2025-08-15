package middleware

import (
	"crypto/sha256"
	"fmt"
	"net/http"
)

// ETag adds ETag header for better caching
func ETag(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip ETag for non-GET requests
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		// Create a response writer that captures the response
		etagWriter := &etagResponseWriter{
			ResponseWriter: w,
			request:        r,
		}

		next.ServeHTTP(etagWriter, r)
	})
}

type etagResponseWriter struct {
	http.ResponseWriter
	body    []byte
	request *http.Request
}

func (w *etagResponseWriter) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)

	// Update ETag with content if headers haven't been written yet
	if w.Header().Get("ETag") == "" {
		hash := sha256.Sum256(w.body)
		etag := fmt.Sprintf(`"%x"`, hash[:8])
		w.Header().Set("ETag", etag)
	}

	return w.ResponseWriter.Write(data)
}

func (w *etagResponseWriter) WriteHeader(statusCode int) {
	// Generate ETag BEFORE calling the original WriteHeader
	if w.Header().Get("ETag") == "" {
		// Use path-based ETag if no content yet, or content-based if available
		var etag string
		if len(w.body) > 0 {
			hash := sha256.Sum256(w.body)
			etag = fmt.Sprintf(`"%x"`, hash[:8])
		} else {
			// Fallback to path-based ETag
			hash := sha256.Sum256([]byte(w.request.URL.Path))
			etag = fmt.Sprintf(`"%x"`, hash[:8])
		}
		w.Header().Set("ETag", etag)
	}

	w.ResponseWriter.WriteHeader(statusCode)
}

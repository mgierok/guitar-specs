package middleware

import (
	"crypto/sha256"
	"fmt"
	"net/http"
)

// ETag adds ETag headers to HTTP responses for improved caching.
// This middleware generates ETags based on response content or request path,
// enabling clients to send conditional requests with If-None-Match headers.
func ETag(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ETags are only meaningful for GET requests that return content
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		// Create a response writer that captures the response content
		// so we can generate an ETag based on the actual response
		etagWriter := &etagResponseWriter{
			ResponseWriter: w,
			request:        r,
		}

		next.ServeHTTP(etagWriter, r)
	})
}

// etagResponseWriter wraps the original ResponseWriter to capture response content
// and generate ETag headers at the appropriate time in the response lifecycle.
type etagResponseWriter struct {
	http.ResponseWriter
	body    []byte        // Accumulates response content for ETag generation
	request *http.Request // Original request for fallback ETag generation
}

// Write captures response data and generates ETag if not already set.
// This ensures ETags are based on the actual response content.
func (w *etagResponseWriter) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)

	// Generate ETag based on content if headers haven't been written yet
	// This allows for content-based ETags which are more accurate
	if w.Header().Get("ETag") == "" {
		hash := sha256.Sum256(w.body)
		etag := fmt.Sprintf(`"%x"`, hash[:8])
		w.Header().Set("ETag", etag)
	}

	return w.ResponseWriter.Write(data)
}

// WriteHeader generates ETag before calling the original WriteHeader.
// This ensures ETags are set before the response is committed to the client.
func (w *etagResponseWriter) WriteHeader(statusCode int) {
	// Generate ETag if not already set by Write method
	if w.Header().Get("ETag") == "" {
		var etag string
		if len(w.body) > 0 {
			// Prefer content-based ETag for accuracy
			hash := sha256.Sum256(w.body)
			etag = fmt.Sprintf(`"%x"`, hash[:8])
		} else {
			// Fallback to path-based ETag when no content is available
			// This ensures ETags are always present for caching
			hash := sha256.Sum256([]byte(w.request.URL.Path))
			etag = fmt.Sprintf(`"%x"`, hash[:8])
		}
		w.Header().Set("ETag", etag)
	}

	w.ResponseWriter.WriteHeader(statusCode)
}

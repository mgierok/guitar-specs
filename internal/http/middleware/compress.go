package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

// Compress adds gzip compression to HTTP responses.
// This middleware compresses responses for supported content types
// when the client indicates support for gzip compression.
func Compress(level int, contentTypes ...string) func(http.Handler) http.Handler {
	// Create a map for efficient content type checking
	allowedTypes := make(map[string]bool)
	for _, ct := range contentTypes {
		allowedTypes[ct] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if client supports gzip compression
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			// Check if content type should be compressed
			contentType := w.Header().Get("Content-Type")
			if contentType == "" {
				// If content type is not set yet, we'll check after the response
				// by wrapping the response writer
				compressWriter := &compressResponseWriter{
					ResponseWriter: w,
					allowedTypes:   allowedTypes,
					level:          level,
				}
				next.ServeHTTP(compressWriter, r)
				compressWriter.Close()
				return
			}

			// Check if current content type should be compressed
			if !shouldCompress(contentType, allowedTypes) {
				next.ServeHTTP(w, r)
				return
			}

			// Compress the response
			gw, err := gzip.NewWriterLevel(w, level)
			if err != nil {
				// If compression fails, serve uncompressed
				next.ServeHTTP(w, r)
				return
			}
			defer gw.Close()

			// Set compression headers
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Add("Vary", "Accept-Encoding")

			// Create a response writer that compresses the content
			compressWriter := &gzipResponseWriter{
				ResponseWriter: w,
				gzipWriter:     gw,
			}

			next.ServeHTTP(compressWriter, r)
		})
	}
}

// shouldCompress checks if a content type should be compressed.
func shouldCompress(contentType string, allowedTypes map[string]bool) bool {
	// Check exact match first
	if allowedTypes[contentType] {
		return true
	}

	// Check prefix matches (e.g., "text/" for "text/html")
	for allowedType := range allowedTypes {
		if strings.HasPrefix(contentType, allowedType) {
			return true
		}
	}

	return false
}

// gzipResponseWriter wraps the response writer to compress content.
type gzipResponseWriter struct {
	http.ResponseWriter
	gzipWriter *gzip.Writer
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	return w.gzipWriter.Write(data)
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

// compressResponseWriter handles compression when content type is not known upfront.
type compressResponseWriter struct {
	http.ResponseWriter
	allowedTypes   map[string]bool
	level          int
	shouldCompress bool
	gzipWriter     *gzip.Writer
	buffer         []byte // Buffer to store data until we know if we should compress
}

func (w *compressResponseWriter) Write(data []byte) (int, error) {
	if !w.shouldCompress {
		// If we shouldn't compress, write directly to the original writer
		return w.ResponseWriter.Write(data)
	}

	// If we should compress but gzipWriter is not initialized yet, buffer the data
	if w.gzipWriter == nil {
		w.buffer = append(w.buffer, data...)
		return len(data), nil
	}

	// Write to gzip writer
	return w.gzipWriter.Write(data)
}

func (w *compressResponseWriter) WriteHeader(statusCode int) {
	contentType := w.Header().Get("Content-Type")
	w.shouldCompress = shouldCompress(contentType, w.allowedTypes)

	if w.shouldCompress {
		// Initialize gzip writer
		gw, err := gzip.NewWriterLevel(w.ResponseWriter, w.level)
		if err == nil {
			w.gzipWriter = gw
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Add("Vary", "Accept-Encoding")

			// Write any buffered data
			if len(w.buffer) > 0 {
				w.gzipWriter.Write(w.buffer)
				w.buffer = nil // Clear buffer
			}
		} else {
			// If compression fails, fall back to uncompressed
			w.shouldCompress = false
			// Write any buffered data directly
			if len(w.buffer) > 0 {
				w.ResponseWriter.Write(w.buffer)
				w.buffer = nil
			}
		}
	} else {
		// Write any buffered data directly if we're not compressing
		if len(w.buffer) > 0 {
			w.ResponseWriter.Write(w.buffer)
			w.buffer = nil
		}
	}

	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *compressResponseWriter) Close() {
	if w.gzipWriter != nil {
		w.gzipWriter.Close()
	}
}

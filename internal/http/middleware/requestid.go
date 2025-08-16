package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// RequestID adds a unique request identifier to each HTTP request.
// This middleware generates a random 16-byte hex string for request tracing
// and debugging purposes.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request already has an ID (e.g., from upstream proxy)
		if r.Header.Get("X-Request-ID") == "" {
			// Generate a new request ID
			id := generateRequestID()
			r.Header.Set("X-Request-ID", id)
		}

		// Add request ID to response headers for client reference
		w.Header().Set("X-Request-ID", r.Header.Get("X-Request-ID"))

		next.ServeHTTP(w, r)
	})
}

// generateRequestID creates a random 16-byte hex string for request identification.
// This provides sufficient uniqueness for request tracing while keeping the ID short.
func generateRequestID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

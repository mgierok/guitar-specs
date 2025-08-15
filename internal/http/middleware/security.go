package middleware

import (
	"net/http"
)

// SecurityHeaders adds security-related HTTP headers to all responses.
// This middleware implements defence-in-depth by setting multiple security headers
// that protect against common web vulnerabilities.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent clickjacking attacks by disallowing frame embedding
		w.Header().Set("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing which can lead to XSS attacks
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Enable legacy XSS protection for older browsers
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Control referrer information leakage to third-party sites
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Tightened Content Security Policy to restrict resource loading
		// This prevents XSS, data injection, and unauthorised resource loading
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self'; "+
				"style-src 'self'; "+
				"img-src 'self' data:; "+
				"font-src 'self'; "+
				"object-src 'none'; "+
				"base-uri 'self'; "+
				"frame-ancestors 'none'")

		// Restrict access to browser APIs that could be abused
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Note: HSTS header is now handled by dedicated HSTS middleware
		// This prevents duplication and allows for better configuration control

		next.ServeHTTP(w, r)
	})
}

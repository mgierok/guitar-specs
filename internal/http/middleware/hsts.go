package middleware

import (
	"net/http"
)

// HSTS adds HTTP Strict Transport Security headers to enforce HTTPS usage.
// This middleware prevents protocol downgrade attacks and ensures all future
// requests from the client use HTTPS, even if the user types http:// in the browser.
func HSTS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only add HSTS header for HTTPS connections
		if r.TLS != nil {
			// max-age=31536000 = 1 year
			// includeSubDomains = applies to all subdomains
			// preload = allows inclusion in browser HSTS preload lists
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}
		
		next.ServeHTTP(w, r)
	})
}


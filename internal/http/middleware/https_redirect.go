package middleware

import (
	"net/http"
)

// HTTPSRedirect redirects HTTP requests to HTTPS when enabled.
// This middleware ensures all traffic uses secure connections and prevents
// protocol downgrade attacks by enforcing HTTPS-only access.
func HTTPSRedirect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request is already using HTTPS
		if r.TLS != nil {
			// Request is already HTTPS, proceed normally
			next.ServeHTTP(w, r)
			return
		}
		
		// Request is HTTP, redirect to HTTPS
		// Preserve the original path and query parameters
		httpsURL := "https://" + r.Host + r.RequestURI
		
		// Use 301 (permanent) redirect for better SEO and caching
		// This tells search engines and browsers that the resource has permanently moved
		http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
	})
}


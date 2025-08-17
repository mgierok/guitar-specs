package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
)

// SecurityHeaders adds security-related HTTP headers to all responses.
// This middleware implements defence-in-depth by setting multiple security headers
// that protect against common web vulnerabilities. It also injects a per-request
// CSP nonce for safe inline/templated scripts.
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

		// Generate CSP nonce
		var nonceBytes [16]byte
		_, _ = rand.Read(nonceBytes[:])
		nonce := base64.StdEncoding.EncodeToString(nonceBytes[:])

		// Content Security Policy with nonce for scripts
		csp := "default-src 'self'; " +
			"script-src 'self' 'nonce-" + nonce + "'; " +
			"style-src 'self'; " +
			"img-src 'self' data:; " +
			"font-src 'self'; " +
			"object-src 'none'; " +
			"base-uri 'self'; " +
			"frame-ancestors 'none'"
		w.Header().Set("Content-Security-Policy", csp)

		// Restrict access to browser APIs that could be abused
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Attach nonce to context so templates can access it
		r = r.WithContext(WithCSPNonce(r.Context(), nonce))

		// Note: HSTS is handled by Cloudflare CDN layer
		next.ServeHTTP(w, r)
	})
}

// context key for CSP nonce
type cspNonceKey struct{}

// WithCSPNonce stores a CSP nonce in the context.
func WithCSPNonce(ctx context.Context, nonce string) context.Context {
	return context.WithValue(ctx, cspNonceKey{}, nonce)
}

// CSPNonceFromContext retrieves a CSP nonce from the context.
func CSPNonceFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(cspNonceKey{})
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

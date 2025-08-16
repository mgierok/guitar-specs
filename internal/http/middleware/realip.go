package middleware

import (
	"net"
	"net/http"
	"strings"
)

// RealIP extracts the real client IP address from proxy headers.
// This middleware handles common proxy scenarios and ensures accurate client IP logging.
func RealIP(trustedProxies []string) func(http.Handler) http.Handler {
	// Convert trusted proxies to net.IP for efficient comparison
	trustedIPs := make([]net.IP, 0, len(trustedProxies))
	for _, proxy := range trustedProxies {
		if ip := net.ParseIP(proxy); ip != nil {
			trustedIPs = append(trustedIPs, ip)
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract real IP from various proxy headers
			realIP := extractRealIP(r, trustedIPs)

			// Set the real IP in the request context for downstream handlers
			r.RemoteAddr = realIP

			next.ServeHTTP(w, r)
		})
	}
}

// extractRealIP determines the real client IP by checking proxy headers in order of preference.
// It validates that the IP comes from a trusted proxy to prevent IP spoofing attacks.
func extractRealIP(r *http.Request, trustedIPs []net.IP) string {
	// First, check if the direct connection IP is trusted
	directIP := extractIPFromAddr(r.RemoteAddr)
	if !isTrustedProxy(directIP, trustedIPs) {
		// If direct connection is not from trusted proxy, don't trust any headers
		return r.RemoteAddr
	}

	// Check X-Forwarded-For header (most common)
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			clientIP := strings.TrimSpace(ips[0])
			if ip := net.ParseIP(clientIP); ip != nil {
				return clientIP
			}
		}
	}

	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		if ip := net.ParseIP(realIP); ip != nil {
			return realIP
		}
	}

	// Check X-Client-IP header
	if clientIP := r.Header.Get("X-Client-IP"); clientIP != "" {
		if ip := net.ParseIP(clientIP); ip != nil {
			return clientIP
		}
	}

	// Check CF-Connecting-IP header (Cloudflare)
	if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
		if ip := net.ParseIP(cfIP); ip != nil {
			return cfIP
		}
	}

	// Fall back to the direct connection IP
	return r.RemoteAddr
}

// extractIPFromAddr extracts the IP address from a network address string.
func extractIPFromAddr(addr string) net.IP {
	// Remove port if present
	if colonIndex := strings.LastIndex(addr, ":"); colonIndex != -1 {
		addr = addr[:colonIndex]
	}
	return net.ParseIP(addr)
}

// isTrustedProxy checks if an IP address is in the list of trusted proxies.
func isTrustedProxy(ip net.IP, trustedIPs []net.IP) bool {
	if ip == nil {
		return false
	}

	for _, trustedIP := range trustedIPs {
		if ip.Equal(trustedIP) {
			return true
		}
	}
	return false
}

// isPrivateIP checks if an IP address is in a private range.
// This helps prevent IP spoofing by rejecting private IPs from untrusted sources.
func isPrivateIP(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		return ip4[0] == 10 || // 10.0.0.0/8
			(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) || // 172.16.0.0/12
			(ip4[0] == 192 && ip4[1] == 168) // 192.168.0.0/16
	}
	return false
}

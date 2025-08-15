package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a sliding window rate limiting mechanism.
// It tracks requests per client IP address and enforces limits within a time window.
type RateLimiter struct {
	requests map[string][]time.Time // Maps client IP to slice of request timestamps
	mu       sync.RWMutex           // Protects concurrent access to requests map
	limit    int                    // Maximum number of requests allowed per window
	window   time.Duration          // Time window for rate limiting
}

// NewRateLimiter creates a new rate limiter with specified limit and window.
// The rate limiter will allow up to 'limit' requests within each 'window' duration.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// RateLimit returns a middleware that enforces rate limiting on HTTP requests.
// It uses a sliding window approach where old requests outside the window are cleaned up
// and new requests are only allowed if they don't exceed the limit.
func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use client IP address as the key for rate limiting
		key := r.RemoteAddr

		rl.mu.Lock()
		now := time.Now()

		// Clean up old requests that are outside the current time window
		// This prevents memory leaks and ensures accurate rate limiting
		if times, exists := rl.requests[key]; exists {
			var valid []time.Time
			for _, t := range times {
				if now.Sub(t) <= rl.window {
					valid = append(valid, t)
				}
			}
			rl.requests[key] = valid
		}

		// Check if the client has exceeded their rate limit
		if len(rl.requests[key]) >= rl.limit {
			rl.mu.Unlock()
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		// Record the current request timestamp for this client
		rl.requests[key] = append(rl.requests[key], now)
		rl.mu.Unlock()

		// Allow the request to proceed
		next.ServeHTTP(w, r)
	})
}

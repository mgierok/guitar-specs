package middleware

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use client IP as key
		key := r.RemoteAddr

		rl.mu.Lock()
		now := time.Now()

		// Clean old requests outside the window
		if times, exists := rl.requests[key]; exists {
			var valid []time.Time
			for _, t := range times {
				if now.Sub(t) <= rl.window {
					valid = append(valid, t)
				}
			}
			rl.requests[key] = valid
		}

		// Check if limit exceeded
		if len(rl.requests[key]) >= rl.limit {
			rl.mu.Unlock()
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		// Add current request
		rl.requests[key] = append(rl.requests[key], now)
		rl.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

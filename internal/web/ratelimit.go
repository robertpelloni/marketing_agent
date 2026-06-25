package web

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// rateLimiter defines a basic IP-based rate limiting middleware.
type rateLimiter struct {
	mu      sync.Mutex
	limiters map[string]*rate.Limiter
	r       rate.Limit
	b       int
}

// newRateLimiter creates a new rate limiter allowing r events per second with a burst of b.
func newRateLimiter(r rate.Limit, b int) *rateLimiter {
	return &rateLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        r,
		b:        b,
	}
}

// getLimiter returns the rate limiter for a specific IP.
func (rl *rateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.r, rl.b)
		rl.limiters[ip] = limiter
	}

	return limiter
}

// middleware provides the rate limiting handler.
func (rl *rateLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use RemoteAddr, in a real proxy scenario this should parse X-Forwarded-For
		ip := r.RemoteAddr
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

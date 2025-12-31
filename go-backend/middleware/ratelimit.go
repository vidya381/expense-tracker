package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// IPRateLimiter tracks rate limiters per IP address
type IPRateLimiter struct {
	ips        map[string]*rate.Limiter
	lastAccess map[string]time.Time
	mu         *sync.RWMutex
	r          rate.Limit
	b          int
}

// NewIPRateLimiter creates a new IP-based rate limiter
// r = requests per second, b = burst size
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips:        make(map[string]*rate.Limiter),
		lastAccess: make(map[string]time.Time),
		mu:         &sync.RWMutex{},
		r:          r,
		b:          b,
	}
}

// GetLimiter returns the rate limiter for the given IP
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
	}

	// Track last access time
	i.lastAccess[ip] = time.Now()

	return limiter
}

// CleanupOldEntries removes rate limiters that haven't been used recently
func (i *IPRateLimiter) CleanupOldEntries() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			i.mu.Lock()
			// Remove entries not accessed in last 5 minutes
			cutoff := time.Now().Add(-5 * time.Minute)
			for ip, lastAccess := range i.lastAccess {
				if lastAccess.Before(cutoff) {
					delete(i.ips, ip)
					delete(i.lastAccess, ip)
				}
			}
			i.mu.Unlock()
		}
	}()
}

// RateLimitMiddleware creates a middleware that limits requests per IP
func RateLimitMiddleware(limiter *IPRateLimiter) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			// Extract IP without port
			if idx := len(ip) - 1; idx >= 0 {
				for i := idx; i >= 0; i-- {
					if ip[i] == ':' {
						ip = ip[:i]
						break
					}
				}
			}

			limiter := limiter.GetLimiter(ip)
			if !limiter.Allow() {
				http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
				return
			}

			next(w, r)
		}
	}
}

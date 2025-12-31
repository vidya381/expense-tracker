package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/vidya381/expense-tracker-backend/utils"
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

// getClientIP extracts the real client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (set by proxies/load balancers)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
		// Take the first one (original client IP)
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ip != "" {
				return ip
			}
		}
	}

	// Fall back to X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// Extract IP without port
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// RateLimitMiddleware creates a middleware that limits requests per IP
func RateLimitMiddleware(limiter *IPRateLimiter) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)

			limiter := limiter.GetLimiter(ip)
			if !limiter.Allow() {
				utils.RespondWithError(w, http.StatusTooManyRequests, "Rate limit exceeded. Please try again later.")
				return
			}

			next(w, r)
		}
	}
}

package middleware

import (
	"net/http"
	"os"
)

// SecurityHeaders adds security-related HTTP headers to responses
func SecurityHeaders(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Prevent clickjacking attacks
		w.Header().Set("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Enable XSS filter in browsers
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Control referrer information
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Restrict browser features
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Content Security Policy for API responses
		// Strict policy since this is a JSON API, not serving HTML
		csp := getCSPHeader()
		w.Header().Set("Content-Security-Policy", csp)

		next(w, r)
	}
}

// getCSPHeader returns the Content Security Policy header value
// Can be configured via environment variable for different environments
func getCSPHeader() string {
	// Check if custom CSP is provided via environment
	customCSP := os.Getenv("CSP_HEADER")
	if customCSP != "" {
		return customCSP
	}

	// Default strict CSP for JSON API
	// This prevents any inline scripts, styles, or external resource loading
	return "default-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'"
}

// StrictTransportSecurity adds HSTS header for HTTPS enforcement
// Only call this in production with HTTPS enabled
func StrictTransportSecurity(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Enable HSTS with 1 year max-age, including subdomains
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		next(w, r)
	}
}

// RequireHTTPS enforces HTTPS in production environments
// Redirects HTTP requests to HTTPS and adds HSTS header
func RequireHTTPS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if HTTPS enforcement is enabled via environment variable
		enforceHTTPS := os.Getenv("ENFORCE_HTTPS")
		if enforceHTTPS != "true" {
			// HTTPS not enforced, continue normally
			next(w, r)
			return
		}

		// Check if request is already HTTPS
		// Check multiple ways since it might be behind a proxy/load balancer
		isHTTPS := r.TLS != nil ||
			r.Header.Get("X-Forwarded-Proto") == "https" ||
			r.Header.Get("X-Forwarded-Ssl") == "on" ||
			r.URL.Scheme == "https"

		if !isHTTPS {
			// Redirect HTTP to HTTPS
			httpsURL := "https://" + r.Host + r.RequestURI
			http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
			return
		}

		// Add HSTS header for HTTPS requests
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		next(w, r)
	}
}

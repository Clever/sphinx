package handlers

import (
	"github.com/Clever/sphinx/ratelimiter"
	"net/http"
)

// START OMIT

// NewHTTPLimiter returns an http.Handler that rate limits and proxies requests.
func NewHTTPLimiter(rateLimiter ratelimiter.RateLimiter, proxy http.Handler) http.Handler {
	return &httpRateLimiter{rateLimiter: rateLimiter, proxy: proxy}
}

// NewHTTPLogger returns an http.Handler that logs the results of rate limiting requests, but
// actually proxies everything.
func NewHTTPLogger(rateLimiter ratelimiter.RateLimiter, proxy http.Handler) http.Handler {
	return &httpRateLogger{rateLimiter: rateLimiter, proxy: proxy}
}

// END OMIT

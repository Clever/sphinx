package handlers

import (
	"github.com/Clever/leakybucket"
	"github.com/Clever/sphinx/common"
	"github.com/Clever/sphinx/ratelimiter"
	"log"
	"net/http"
)

func addRateLimitHeaders(w http.ResponseWriter, statuses []ratelimiter.Status) {}

// START OMIT
type httpRateLimiter struct {
	rateLimiter ratelimiter.RateLimiter
	proxy       http.Handler
}

func (hrl httpRateLimiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request := common.HTTPToSphinxRequest(r)
	matches, err := hrl.rateLimiter.Add(request)
	if err != nil && err != leakybucket.ErrorFull {
		log.Printf("ERROR: %s\n", err)
		w.WriteHeader(500)
		return
	}

	// Write the rate limiter status to the response header
	addRateLimitHeaders(w, matches)
	if err == leakybucket.ErrorFull {
		w.WriteHeader(429)
		return
	}
	hrl.proxy.ServeHTTP(w, r)
}

// END OMIT

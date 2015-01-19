package handlers

import (
	"github.com/Clever/leakybucket"
	"github.com/Clever/sphinx/common"
	"github.com/Clever/sphinx/ratelimiter"
	"log"
	"net/http"
)

func stringifyLimitHeaders(statuses []ratelimiter.Status) string { return "" }

// START OMIT
type httpRateLogger httpRateLimiter

func (hrl httpRateLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request := common.HTTPToSphinxRequest(r)

	matches, err := hrl.rateLimiter.Add(request)
	if err != nil && err != leakybucket.ErrorFull {
		log.Printf("ERROR: %s", err)
		hrl.proxy.ServeHTTP(w, r)
		return
	}

	// Log the rate limit hearders
	log.Printf("RATE LIMIT HEADERS: %s\n", stringifyLimitHeaders(matches))
	if err == leakybucket.ErrorFull {
		// Log when the bucket is full
		log.Printf("BUCKET FULL")
	}
	// Always proxy the request
	hrl.proxy.ServeHTTP(w, r)
}

// END OMIT

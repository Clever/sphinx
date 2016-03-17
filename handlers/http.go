package handlers

import (
	"github.com/Clever/leakybucket"
	"github.com/Clever/sphinx/common"
	"github.com/Clever/sphinx/ratelimiter"
	"github.com/pborman/uuid"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	// StatusTooManyRequests represents HTTP 429, missing from net/http
	StatusTooManyRequests = 429 // not in net/http package
)

func flattenRateLimitHeaders(headers http.Header) common.M {
	flatHeaders := common.M{}
	for key, vals := range headers {
		flatHeaders[key] = strings.Join(vals, ";")
	}
	return flatHeaders
}

type httpRateLimiter struct {
	rateLimiter  ratelimiter.RateLimiter
	proxy        http.Handler
	allowOnError bool // Do not limit on errors when true
}

func (hrl httpRateLimiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	guid := uuid.New()
	request := common.HTTPToSphinxRequest(r)
	matches, err := hrl.rateLimiter.Add(request)
	switch {
	case err == leakybucket.ErrorFull:
		addRateLimitHeaders(w, matches)
		common.Log.InfoD("request-finished", common.ConcatWithRequest(
			common.M{"guid": guid, "limit": true}, request))
		w.WriteHeader(StatusTooManyRequests)
	case err != nil && hrl.allowOnError:
		common.Log.WarnD("request-finished", common.ConcatWithRequest(
			common.M{"guid": guid, "err": err}, request))
		addRateLimitHeaders(w, []ratelimiter.Status{ratelimiter.NilStatus})
		hrl.proxy.ServeHTTP(w, r)
	case err != nil:
		common.Log.ErrorD("request-finished",
			common.ConcatWithRequest(
				common.M{"guid": guid,
					"err": err}, request))
		w.WriteHeader(http.StatusInternalServerError)

	default:
		common.Log.InfoD("request-finished", common.ConcatWithRequest(common.M{"guid": guid, "limit": false}, request))
		addRateLimitHeaders(w, matches)
		hrl.proxy.ServeHTTP(w, r)
	}
}

type httpRateLogger httpRateLimiter

func (hrl httpRateLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	guid := uuid.New()
	request := common.HTTPToSphinxRequest(r)
	matches, err := hrl.rateLimiter.Add(request)
	if err != nil && err != leakybucket.ErrorFull {
		log.Printf("[%s] ERROR: %s", guid, err)
		hrl.proxy.ServeHTTP(w, r)
		return
	}

	rateLimitResponse := flattenRateLimitHeaders(getRateLimitHeaders(matches))
	if err == leakybucket.ErrorFull {
		rateLimitResponse["limit"] = true
	} else {
		rateLimitResponse["limit"] = false
	}
	rateLimitResponse["guid"] = guid

	common.Log.InfoD("http-logger", common.ConcatWithRequest(rateLimitResponse, request))
	hrl.proxy.ServeHTTP(w, r)
}

func uintToString(num uint) string {
	return strconv.Itoa(int(num))
}

func int64ToString(num int64) string {
	return strconv.Itoa(int(num))
}

func initHeaders() map[string][]string {
	headers := map[string][]string{}
	for _, header := range []string{"Limit", "Reset", "Remaining", "Bucket"} {
		headerName := "X-Ratelimit-" + header
		if headers[headerName] == nil {
			headers[headerName] = []string{}
		}
	}
	return headers
}

func getRateLimitHeaders(statuses []ratelimiter.Status) map[string][]string {
	if len(statuses) == 0 {
		return map[string][]string{}
	}
	headers := initHeaders()
	for _, status := range statuses {
		headers["X-Ratelimit-Limit"] = append(headers["X-Ratelimit-Limit"], uintToString(status.Capacity))
		headers["X-Ratelimit-Reset"] = append(headers["X-Ratelimit-Reset"], int64ToString(status.Reset.Unix()))
		headers["X-Ratelimit-Remaining"] = append(headers["X-Ratelimit-Remaining"], uintToString(status.Remaining))
		headers["X-Ratelimit-Bucket"] = append(headers["X-Ratelimit-Bucket"], status.Name)
	}
	return headers
}

func addRateLimitHeaders(w http.ResponseWriter, statuses []ratelimiter.Status) {
	for header, values := range getRateLimitHeaders(statuses) {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}
}

// NewHTTPLimiter returns an http.Handler that rate limits and proxies requests.
func NewHTTPLimiter(rateLimiter ratelimiter.RateLimiter, proxy http.Handler, allowOnError bool) http.Handler {
	return &httpRateLimiter{rateLimiter: rateLimiter, proxy: proxy, allowOnError: allowOnError}
}

// NewHTTPLogger returns an http.Handler that logs the results of rate limiting requests, but
// actually proxies everything.
func NewHTTPLogger(rateLimiter ratelimiter.RateLimiter, proxy http.Handler) http.Handler {
	return &httpRateLogger{rateLimiter: rateLimiter, proxy: proxy}
}

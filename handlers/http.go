package http

import (
	"github.com/Clever/leakybucket"
	"github.com/Clever/sphinx"
	"net/http"
	"net/http/httputil"
	"strconv"
)

func parseRequest(r *http.Request) sphinx.Request {
	return map[string]interface{}{
		"path":    r.URL.Path,
		"headers": r.Header,
	}
}

type HTTPRateLimiter struct {
	ratelimiter sphinx.RateLimiter
	proxy       *httputil.ReverseProxy
}

func (hrl HTTPRateLimiter) Handle(w http.ResponseWriter, r *http.Request) {
	matches, err := hrl.ratelimiter.Add(parseRequest(r))
	if err != nil && err != leakybucket.ErrorFull {
		// TODO: Send to sentry.
		w.WriteHeader(500)
		return
	}
	addRateLimitHeaders(w, matches)
	if err == leakybucket.ErrorFull {
		w.WriteHeader(429)
		return
	}
	hrl.proxy.ServeHTTP(w, r)
}

func uintToString(num uint) string {
	return strconv.Itoa(int(num))
}

func int64ToString(num int64) string {
	return strconv.Itoa(int(num))
}

func addRateLimitHeaders(w http.ResponseWriter, statuses []sphinx.Status) {
	for _, status := range statuses {
		w.Header().Add("X-Rate-Limit-Limit", uintToString(status.Capacity))
		w.Header().Add("X-Rate-Limit-Reset", int64ToString(status.Reset.Unix()))
		w.Header().Add("X-Rate-Limit-Remaining", uintToString(status.Remaining))
		w.Header().Add("X-Rate-Limit-Bucket", status.Name)
	}
}

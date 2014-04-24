package http

import (
	"github.com/Clever/leakybucket"
	"github.com/Clever/sphinx"
	"github.com/Clever/sphinx/common"
	"log"
	"net/http"
	"strconv"
)

func parseRequest(r *http.Request) common.Request {
	return map[string]interface{}{
		"path":       r.URL.Path,
		"headers":    r.Header,
		"remoteaddr": r.RemoteAddr,
	}
}

type HTTPRateLimiter struct {
	ratelimiter sphinx.RateLimiter
	proxy       http.Handler
}

func (hrl HTTPRateLimiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	matches, err := hrl.ratelimiter.Add(parseRequest(r))
	if err != nil && err != leakybucket.ErrorFull {
		// TODO: Log to stderr
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

type HTTPRateLogger HTTPRateLimiter

func (hrl HTTPRateLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	matches, err := hrl.ratelimiter.Add(parseRequest(r))
	if err != nil && err != leakybucket.ErrorFull {
		log.Printf("ERROR")
		hrl.proxy.ServeHTTP(w, r)
		return
	}
	log.Printf("RATE LIMIT HEADERS: %#v", getRateLimitHeaders(matches))
	if err == leakybucket.ErrorFull {
		log.Printf("FULL")
	}
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
		headerName := "X-Rate-Limit-" + header
		if headers[headerName] == nil {
			headers[headerName] = []string{}
		}
	}
	return headers
}

func getRateLimitHeaders(statuses []sphinx.Status) map[string][]string {
	if len(statuses) == 0 {
		return map[string][]string{}
	}
	headers := initHeaders()
	for _, status := range statuses {
		headers["X-Rate-Limit-Limit"] = append(headers["X-Rate-Limit-Limit"], uintToString(status.Capacity))
		headers["X-Rate-Limit-Reset"] = append(headers["X-Rate-Limit-Reset"], int64ToString(status.Reset.Unix()))
		headers["X-Rate-Limit-Remaining"] = append(headers["X-Rate-Limit-Remaining"], uintToString(status.Remaining))
		headers["X-Rate-Limit-Bucket"] = append(headers["X-Rate-Limit-Bucket"], status.Name)
	}
	return headers
}

func addRateLimitHeaders(w http.ResponseWriter, statuses []sphinx.Status) {
	for header, values := range getRateLimitHeaders(statuses) {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}
}

func NewHTTPLimiter(ratelimiter sphinx.RateLimiter, proxy http.Handler) HTTPRateLimiter {
	return HTTPRateLimiter{ratelimiter: ratelimiter, proxy: proxy}
}

func NewHTTPLogger(ratelimiter sphinx.RateLimiter, proxy http.Handler) HTTPRateLogger {
	return HTTPRateLogger{ratelimiter: ratelimiter, proxy: proxy}
}

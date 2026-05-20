package handlers

import (
	"encoding/base64"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Clever/leakybucket"
	"github.com/Clever/sphinx/common"
	"github.com/Clever/sphinx/ratelimiter"
	"github.com/launchdarkly/go-sdk-common/v3/ldcontext"
	"github.com/pborman/uuid"
)

// FlagChecker evaluates a LaunchDarkly boolean feature flag.
type FlagChecker interface {
	BoolVariation(key string, context ldcontext.Context, defaultVal bool) (bool, error)
}

// ldContextKey returns the LaunchDarkly context key for the request, or "" if the request
// type is not one we evaluate flags for.
//
// - Basic auth:  decode base64 credential, extract client_id → "app:{client_id}"
// - Edmentum:    X-Edmentum-Rate-Limit header value       → "app:{value}"
// - Bearer:      opaque token, no per-app identity         → "bearer-global"
func ldContextKey(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Basic ") {
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic "))
		if err == nil {
			if parts := strings.SplitN(string(decoded), ":", 2); len(parts) == 2 && parts[0] != "" {
				return "app:" + parts[0]
			}
		}
	}
	if strings.HasPrefix(auth, "Bearer ") {
		return "bearer-global"
	}
	if val := r.Header.Get("X-Edmentum-Rate-Limit"); val != "" {
		return "app:" + val
	}
	return ""
}

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
	ldChecker    FlagChecker
	ldFlagName   string
}

func (hrl httpRateLimiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	guid := r.Header.Get("X-Request-Id")
	if hrl.ldChecker != nil {
		if ctxKey := ldContextKey(r); ctxKey != "" {
			ldCtx := ldcontext.NewBuilder(ctxKey).Build()
			bypass, _ := hrl.ldChecker.BoolVariation(hrl.ldFlagName, ldCtx, false)
			if bypass {
				addRateLimitHeaders(w, []ratelimiter.Status{ratelimiter.BypassStatus})
				hrl.proxy.ServeHTTP(w, r)
				return
			}
		}
	}
	request := common.HTTPToSphinxRequest(r)
	matches, err := hrl.rateLimiter.Add(request)
	if err == leakybucket.ErrorFull {
		addRateLimitHeaders(w, matches)
		common.Log.InfoD("ratelimited", common.ConcatWithRequest(common.M{"guid": guid}, request))
		w.WriteHeader(StatusTooManyRequests)
		return
	}
	if err != nil {
		common.Log.WarnD("error",
			common.ConcatWithRequest(common.M{"guid": guid, "err": err.Error()}, request))
		if !hrl.allowOnError {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		matches = []ratelimiter.Status{ratelimiter.NilStatus}
	}
	addRateLimitHeaders(w, matches)
	hrl.proxy.ServeHTTP(w, r)
}

type httpRateLogger httpRateLimiter

func (hrl httpRateLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	guid := uuid.New()
	if hrl.ldChecker != nil {
		if ctxKey := ldContextKey(r); ctxKey != "" {
			ldCtx := ldcontext.NewBuilder(ctxKey).Build()
			bypass, _ := hrl.ldChecker.BoolVariation(hrl.ldFlagName, ldCtx, false)
			if bypass {
				common.Log.InfoD("http-logger", common.ConcatWithRequest(
					common.M{"guid": guid, "limit": false, "bypass": true}, common.HTTPToSphinxRequest(r),
				))
				hrl.proxy.ServeHTTP(w, r)
				return
			}
		}
	}
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
func NewHTTPLimiter(rateLimiter ratelimiter.RateLimiter, proxy http.Handler, allowOnError bool, ldChecker FlagChecker, ldFlagName string) http.Handler {
	return &httpRateLimiter{rateLimiter: rateLimiter, proxy: proxy, allowOnError: allowOnError, ldChecker: ldChecker, ldFlagName: ldFlagName}
}

// NewHTTPLogger returns an http.Handler that logs the results of rate limiting requests, but
// actually proxies everything.
func NewHTTPLogger(rateLimiter ratelimiter.RateLimiter, proxy http.Handler, ldChecker FlagChecker, ldFlagName string) http.Handler {
	return &httpRateLogger{rateLimiter: rateLimiter, proxy: proxy, ldChecker: ldChecker, ldFlagName: ldFlagName}
}

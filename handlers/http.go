package http

import (
	"github.com/Clever/leakybucket"
	"github.com/Clever/sphinx"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type HTTPRequest struct {
	request *http.Request
}

func (h HTTPRequest) Properties() map[string]interface{} {

}

func NewHTTPHandler(host string, limiter sphinx.RateLimiter) (http.HandlerFunc, err) {
	remote, err := url.Parse(host)
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	return func(w http.ResponseWriter, r *http.Request) {
		buckets, err := limit.Add(HTTPRequest{request: r})
		if err != nil && err == leakybucket.ErrorFull {
			// Send back rate limited response
		} else if err != nil {
			// Send back server error response
		}
		// Set rate limit headers
		proxy.SeverHTTP(w, r)
	}, nil
}

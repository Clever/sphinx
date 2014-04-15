package http

import (
	"github.com/Clever/sphinx"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"
)

func constructMockRequestWithHeaders(headers map[string][]string) *http.Request {
	testUrl, err := url.Parse("https://google.com/trolling/path")
	if err != nil {
		panic(err)
	}

	return &http.Request{
		Header: headers,
		URL:    testUrl,
	}
}

func compareStrings(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Fatalf("expected %s, received %s", expected, actual)
	}
}

func compareHeader(t *testing.T, headers http.Header, key string, values []string) {
	if len(headers[key]) != len(values) {
		t.Fatalf("expected %d '%s' headers, received %d", len(values), len(headers[key]))
	}
	for i, expected := range values {
		compareStrings(t, expected, headers[key][i])
	}
}

type MockRateLimiter struct {
	Mock   mock.Mock
	limits []sphinx.Limit
}

func (r *MockRateLimiter) Configuration() sphinx.Configuration {
	return sphinx.Configuration{}
}

func (r *MockRateLimiter) Limits() []sphinx.Limit {
	return r.limits
}

func (r *MockRateLimiter) Add(request sphinx.Request) ([]sphinx.Status, error) {
	// args := r.Mock.Called(request)
	return nil, nil
}

func (r *MockRateLimiter) SetLimits(limits []sphinx.Limit) {
	r.limits = limits
}

func constructHTTPRateLimiter(ratelimiter *sphinx.RateLimiter, proxy *httputil.ReverseProxy) HTTPRateLimiter {
	return HTTPRateLimiter{ratelimiter: &MockRateLimiter{}}
}

func TestParsesHeaders(t *testing.T) {
	request := parseRequest(constructMockRequestWithHeaders(map[string][]string{
		"Authorization":   []string{"Bearer 12345"},
		"X-Forwarded-For": []string{"IP1", "IP2"},
	}))
	if len(request["headers"].(http.Header)) != 2 {
		t.Fatalf("expected 2 headers, recevied %d", len(request["headers"].(http.Header)))
	}

	compareHeader(t, request["headers"].(http.Header), "Authorization", []string{"Bearer 12345"})
	compareHeader(t, request["headers"].(http.Header), "X-Forwarded-For", []string{"IP1", "IP2"})
	compareStrings(t, request["path"].(string), "/trolling/path")
}

func TestAddHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	statuses := make([]sphinx.Status, 0)
	addRateLimitHeaders(w, statuses)
	compareHeader(t, w.Header(), "X-Rate-Limit-Limit", []string{})
	compareHeader(t, w.Header(), "X-Rate-Limit-Reset", []string{})
	compareHeader(t, w.Header(), "X-Rate-Limit-Remaining", []string{})
	compareHeader(t, w.Header(), "X-Rate-Limit-Bucket", []string{})
}

func TestHandleWhenErr(t *testing.T) {}

func TestHandleWhenFull(t *testing.T) {}

func TestHandleWhenNotFull(t *testing.T) {}

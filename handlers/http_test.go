package http

import (
	"github.com/Clever/sphinx"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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
	*mock.Mock
	limits []sphinx.Limit
}

func (r *MockRateLimiter) Configuration() sphinx.Configuration {
	return sphinx.Configuration{}
}
func (r *MockRateLimiter) Limits() []sphinx.Limit {
	return r.limits
}
func (r *MockRateLimiter) Add(request sphinx.Request) ([]sphinx.Status, error) {
	args := r.Mock.Called(request)
	return args.Get(0).([]sphinx.Status), args.Error(1)
}
func (r *MockRateLimiter) SetLimits(limits []sphinx.Limit) {
	r.limits = limits
}

type MockProxy struct {
	*mock.Mock
}

func (p *MockProxy) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// Calling only for side effects
	_ = p.Mock.Called(rw, r)
}

func constructHTTPRateLimiter() HTTPRateLimiter {
	return HTTPRateLimiter{
		ratelimiter: &MockRateLimiter{Mock: new(mock.Mock)},
		proxy:       &MockProxy{Mock: new(mock.Mock)},
	}
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

func TestHandleWhenNotFull(t *testing.T) {
	limiter := constructHTTPRateLimiter()
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "http://google.com", strings.NewReader("thebody"))
	if err != nil {
		t.Fatal(err)
	}
	limiter.ratelimiter.(*MockRateLimiter).Mock.On("Add", mock.AnythingOfTypeArgument("sphinx.Request")).Return(make([]sphinx.Status, 0), nil).Once()
	limiter.proxy.(*MockProxy).Mock.On("ServeHTTP", w, r).Return().Once()
	limiter.Handle(w, r)
	// Test that headers are correct
}

func TestHandleWhenFull(t *testing.T) {}

func TestHandleWhenErr(t *testing.T) {}

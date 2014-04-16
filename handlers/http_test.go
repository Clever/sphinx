package http

import (
	"errors"
	"fmt"
	"github.com/Clever/leakybucket"
	"github.com/Clever/sphinx"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
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
	if headers[key] == nil {
		t.Fatalf("header %s does not exist", key)
	}
	if len(headers[key]) != len(values) {
		t.Fatalf("expected %d '%s' headers, received %d", len(values), key, len(headers[key]))
	}
	for i, expected := range values {
		compareStrings(t, expected, headers[key][i])
	}
}

func compareStatusesToHeader(t *testing.T, header http.Header, statuses []sphinx.Status) {
	limits := []string{}
	resets := []string{}
	remainings := []string{}
	buckets := []string{}
	for _, status := range statuses {
		limits = append(limits, uintToString(status.Capacity))
		resets = append(resets, int64ToString(status.Reset.Unix()))
		remainings = append(remainings, uintToString(status.Remaining))
		buckets = append(buckets, status.Name)
	}
	compareHeader(t, header, "X-Rate-Limit-Limit", limits)
	compareHeader(t, header, "X-Rate-Limit-Reset", resets)
	compareHeader(t, header, "X-Rate-Limit-Remaining", remainings)
	compareHeader(t, header, "X-Rate-Limit-Bucket", buckets)
}

func assertNoRateLimitHeaders(t *testing.T, header http.Header) {
	for _, key := range []string{"Limit", "Reset", "Remaining", "Bucket"} {
		val := header.Get(fmt.Sprintf("X-Rate-Limit-%s", key))
		if val != "" {
			t.Fatalf("expected nil for header %s, got %#v", key, val)
		}
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

func TestAddHeadersNoStatus(t *testing.T) {
	w := httptest.NewRecorder()
	statuses := []sphinx.Status{}
	addRateLimitHeaders(w, statuses)
	assertNoRateLimitHeaders(t, w.Header())
}

func TestAddHeadersOneStatus(t *testing.T) {
	w := httptest.NewRecorder()
	statuses := []sphinx.Status{
		{Capacity: uint(10), Reset: time.Now(), Remaining: uint(10), Name: "test"},
	}
	addRateLimitHeaders(w, statuses)
	compareStatusesToHeader(t, w.Header(), statuses)
}

func TestAddHeadersMultipleStatus(t *testing.T) {
	w := httptest.NewRecorder()
	statuses := []sphinx.Status{
		{Capacity: uint(10), Reset: time.Now(), Remaining: uint(10), Name: "test1"},
		{Capacity: uint(100), Reset: time.Now(), Remaining: uint(100), Name: "test2"},
		{Capacity: uint(1000), Reset: time.Now(), Remaining: uint(1000), Name: "test3"},
	}
	addRateLimitHeaders(w, statuses)
	compareStatusesToHeader(t, w.Header(), statuses)
}

var anySphinxRequest = mock.AnythingOfTypeArgument("sphinx.Request")
var sphinxStatus = sphinx.Status{
	Capacity:  uint(10),
	Reset:     time.Now(),
	Remaining: uint(10),
	Name:      "test",
}

func TestHandleWhenNotFull(t *testing.T) {
	limiter := constructHTTPRateLimiter()
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "http://google.com", strings.NewReader("thebody"))
	if err != nil {
		t.Fatal(err)
	}
	statuses := []sphinx.Status{sphinxStatus}

	limitMock := limiter.ratelimiter.(*MockRateLimiter).Mock
	limitMock.On("Add", anySphinxRequest).Return(statuses, nil).Once()

	proxyMock := limiter.proxy.(*MockProxy).Mock
	proxyMock.On("ServeHTTP", w, r).Return().Once()

	limiter.Handle(w, r)

	compareStatusesToHeader(t, w.Header(), statuses)
	// commented out until https://github.com/stretchr/testify/issues/31 is resolved
	// limitMock.AssertExpectations(t)
	proxyMock.AssertExpectations(t)
}

func TestHandleWhenFull(t *testing.T) {
	limiter := constructHTTPRateLimiter()
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "http://google.com", strings.NewReader("thebody"))
	if err != nil {
		t.Fatal(err)
	}
	statuses := []sphinx.Status{sphinxStatus}

	limitMock := limiter.ratelimiter.(*MockRateLimiter).Mock
	limitMock.On("Add", anySphinxRequest).Return(statuses, leakybucket.ErrorFull).Once()

	limiter.Handle(w, r)

	compareStatusesToHeader(t, w.Header(), statuses)
	if w.Code != 429 {
		t.Fatalf("expected status 429, received %d", w.Code)
	}
	// commented out until https://github.com/stretchr/testify/issues/31 is resolved
	// limitMock.AssertExpectations(t)
}

func TestHandleWhenErrWithStatus(t *testing.T) {
	limiter := constructHTTPRateLimiter()
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "http://google.com", strings.NewReader("thebody"))
	if err != nil {
		t.Fatal(err)
	}
	statuses := []sphinx.Status{sphinxStatus}

	limitMock := limiter.ratelimiter.(*MockRateLimiter).Mock
	limitMock.On("Add", anySphinxRequest).Return(statuses, errors.New("random error")).Once()

	limiter.Handle(w, r)
	assertNoRateLimitHeaders(t, w.Header())

	if w.Code != 500 {
		t.Fatalf("expected status 500, received %d", w.Code)
	}
	// commented out until https://github.com/stretchr/testify/issues/31 is resolved
	// limitMock.AssertExpectations(t)
}

func TestHandleWhenErrWithoutStatus(t *testing.T) {
	limiter := constructHTTPRateLimiter()
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "http://google.com", strings.NewReader("thebody"))
	if err != nil {
		t.Fatal(err)
	}
	statuses := []sphinx.Status{}

	limitMock := limiter.ratelimiter.(*MockRateLimiter).Mock
	limitMock.On("Add", anySphinxRequest).Return(statuses, errors.New("random error")).Once()

	limiter.Handle(w, r)
	assertNoRateLimitHeaders(t, w.Header())

	if w.Code != 500 {
		t.Fatalf("expected status 500, received %d", w.Code)
	}
	// commented out until https://github.com/stretchr/testify/issues/31 is resolved
	// limitMock.AssertExpectations(t)
}

package sphinx

import (
	"bytes"
	"fmt"
	"github.com/Clever/leakybucket"
	"github.com/Clever/leakybucket/memory"
	"github.com/Clever/sphinx/common"
	"github.com/Clever/sphinx/matchers"
	"strings"
	"testing"
)

func checkStatusForRequests(ratelimiter RateLimiter,
	request common.Request, num int, expectedStatuses []Status) error {

	var statuses []Status
	var err error
	for i := 0; i < num; i++ {
		statuses, err = ratelimiter.Add(request)
		if err != nil {
			return err
		}
	}

	if len(statuses) != len(expectedStatuses) {
		return fmt.Errorf("expected to match %d buckets. Got: %d",
			len(expectedStatuses), len(statuses))
	}
	for i, status := range expectedStatuses {
		if status.Remaining != statuses[i].Remaining && status.Name != statuses[i].Name {
			return fmt.Errorf("expected %d remaining for the %s limit. Found: %d Remaining, %s Limit",
				statuses[i].Remaining, statuses[i].Name, status.Remaining, status.Name)
		}
	}

	return nil
}

// ratelimiter is initialized properly based on config
func TestNewRateLimiter(t *testing.T) {

	config, err := NewConfiguration("./example.yaml")
	if err != nil {
		t.Error("could not load example configuration")
	}

	ratelimiter, err := NewRateLimiter(config)
	if err != nil {
		t.Errorf("Error while instantiating ratelimiter: %s", err.Error())
	}
	if len(ratelimiter.Configuration().Limits) !=
		len(ratelimiter.Limits()) {
		t.Error("expected number of limits in configuration to match instantiated limits")
	}
}

// test that matcher errors are bubbled up
func TestBadConfiguration(t *testing.T) {

	var configBuf = bytes.NewBufferString(`
proxy:
  handler: http
  host: http://proxy.example.com
  listen: :8080
storage:
  type: memory
limits:
  test:
    interval: 15  # in seconds
    max: 200
`)

	// header matchers are verified
	configBuf.WriteString(`
    keys:
      headers:
        - Authorization
    matches:
      headers:
        match_any:
          - "Authorization": "Bearer.*"
          - name: "X-Forwarded-For"
`)
	configuration, err := loadAndValidateConfig(configBuf.Bytes())
	if err != nil {
		t.Error("configuration failed with error", err)
	}

	_, err = NewRateLimiter(configuration)
	if err == nil {
		t.Error("Expected header matcher error, got none")
	} else if !strings.Contains(err.Error(), "InvalidMatcherConfig: headers") {
		t.Errorf("Expected a InvalidMatcherConfig error, got different error: %s", err.Error())
	}

}

// adds different kinds of requests and checks limit Status
// focusses on single bucket adds
func TestSimpleAdd(t *testing.T) {
	config, err := NewConfiguration("./example.yaml")
	if err != nil {
		t.Error("could not load example configuration")
	}
	ratelimiter, err := NewRateLimiter(config)

	request := common.Request{
		"path": "/special/resources/123",
		"headers": common.ConstructMockRequestWithHeaders(map[string][]string{
			"Authorization":   []string{"Bearer 12345"},
			"X-Forwarded-For": []string{"IP1", "IP2"},
		}).Header,
		"remoteaddr": "127.0.0.1",
	}
	if err = checkStatusForRequests(
		ratelimiter, request, 5, []Status{
			Status{Remaining: 195, Name: "bearer-special"}}); err != nil {
		t.Error(err)
	}

	request = common.Request{
		"path": "/resources/123",
		"headers": common.ConstructMockRequestWithHeaders(map[string][]string{
			"Authorization": []string{"Basic 12345"},
		}).Header,
	}

	if err = checkStatusForRequests(
		ratelimiter, request, 1, []Status{
			Status{Remaining: 195, Name: "basic-simple"}}); err != nil {
		t.Error(err)
	}
}

type NeverMatch struct{}

func (m NeverMatch) Match(req common.Request) bool {
	return false
}

func createLimit(numMatchers int) *Limit {
	neverMatchers := []matchers.Matcher{}
	for i := 0; i < numMatchers; i++ {
		neverMatchers = append(neverMatchers, NeverMatch{})
	}
	limit := &Limit{
		matcher: requestMatcher{
			Matches:  neverMatchers,
			Excludes: neverMatchers,
		},
	}
	return limit
}

var benchMatch = func(b *testing.B, numMatchers int) {
	limit := createLimit(numMatchers)
	request := common.Request{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limit.match(request)
	}
}

func BenchmarkMatch1(b *testing.B) {
	benchMatch(b, 1)
}

func BenchmarkMatch100(b *testing.B) {
	benchMatch(b, 100)
}

func createRateLimiter(numLimits int) RateLimiter {
	limit := createLimit(1)
	rateLimiter := &sphinxRateLimiter{}
	limits := []*Limit{}
	for i := 0; i < numLimits; i++ {
		limits = append(limits, limit)
	}
	rateLimiter.limits = limits
	return rateLimiter
}

var benchAdd = func(b *testing.B, numLimits int) {
	rateLimiter := createRateLimiter(numLimits)
	request := common.Request{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rateLimiter.Add(request)
	}
}

func BenchmarkAdd1(b *testing.B) {
	benchAdd(b, 1)
}

func BenchmarkAdd100(b *testing.B) {
	benchAdd(b, 100)
}

// assert that the right bucket keys are generated for requests
func TestLimitKeys(t *testing.T) {
	keys, err := resolveLimitKeys(map[string]interface{}{
		"headers": []string{"Authorization", "X-Forwarded-For"},
		"ip":      []string{""},
	})
	if err != nil {
		t.Errorf("Error while creating limitkeys for test", err)
	}
	limit := Limit{
		Name: "test-limit",
		keys: keys,
	}

	request := common.Request{
		"path": "/resources/123",
		"headers": common.ConstructMockRequestWithHeaders(map[string][]string{
			"Authorization": []string{"Basic 12345"},
		}).Header,
	}
	if limit.bucketName(request) != "test-limit-Authorization:Basic 12345" {
		t.Errorf("Invalid bucketname for test-limit: %s",
			limit.bucketName(request))
	}
}

// limit.bucketName creates compound keys from multiple limitkeys
func TestCompoundLimitKeys(t *testing.T) {
	keys, err := resolveLimitKeys(map[string]interface{}{
		"headers": []string{"Authorization", "X-Forwarded-For"},
		"ip":      []string{""},
	})
	if err != nil {
		t.Errorf("Error while creating limitkeys for test", err)
	}
	limit := Limit{
		Name: "test-limit",
		keys: keys,
	}

	request := common.Request{
		"path":       "/resources/123",
		"remoteaddr": "127.0.0.1",
		"headers": common.ConstructMockRequestWithHeaders(map[string][]string{
			"Authorization":   []string{"Basic 12345"},
			"X-Forwarded-For": []string{"192.0.0.1"},
		}).Header,
	}
	if limit.bucketName(request) !=
		"test-limit-Authorization:Basic 12345-X-Forwarded-For:192.0.0.1-ip:127.0.0.1" {
		t.Errorf("Invalid compound bucketname for test-limit: %s",
			limit.bucketName(request))
	}
}

// limit.BucketName works when headers are empty
func TestLimitKeyWithEmptyHeaders(t *testing.T) {
	keys, err := resolveLimitKeys(map[string]interface{}{
		"headers": []string{"Authorization", "X-Forwarded-For"},
		"ip":      []string{""},
	})
	if err != nil {
		t.Errorf("Error while creating limitkeys for test", err)
	}
	limit := Limit{
		Name: "test-limit",
		keys: keys,
	}

	request := common.Request{
		"path":       "/resources/123",
		"remoteaddr": "127.0.0.1",
		"headers":    common.ConstructMockRequestWithHeaders(map[string][]string{}).Header,
	}
	if limit.bucketName(request) !=
		"test-limit-ip:127.0.0.1" {
		t.Errorf("Invalid bucketname with no headers for test-limit: %s",
			limit.bucketName(request))
	}
	request = common.Request{
		"path":    "/resources/123",
		"headers": common.ConstructMockRequestWithHeaders(map[string][]string{}).Header,
	}
	if limit.bucketName(request) !=
		"test-limit-" {
		t.Errorf("Invalid bucketname with no valid request data for test-limit: %s",
			limit.bucketName(request))
	}
}

// ensures that Limit.Match exhibits expected behavior
func TestLimitMatch(t *testing.T) {
	config, err := NewConfiguration("./example.yaml")
	if err != nil {
		t.Error("could not load example configuration")
	}

	// matches name: Authorization, match: bearer (from example.yaml)
	matchers, err := resolveMatchers(config.Limits["basic-simple"].Matches)
	// excludes path: /special/resoures/.*
	excludes, err := resolveMatchers(config.Limits["basic-simple"].Excludes)

	limit := Limit{
		Name: "test-limit",
		matcher: requestMatcher{
			Matches:  matchers,
			Excludes: excludes,
		},
	}

	request := common.Request{
		"path": "/resources/123",
		"headers": common.ConstructMockRequestWithHeaders(map[string][]string{
			"Authorization": []string{"Basic 12345"},
		}).Header,
	}
	if !limit.match(request) {
		t.Error("Expected basic-easy to match request")
	}

	request = common.Request{
		"path": "/special/resources/123",
		"headers": common.ConstructMockRequestWithHeaders(map[string][]string{
			"Authorization": []string{"Basic 12345"},
		}).Header,
	}
	if limit.match(request) {
		t.Error("Request with Excludes path should NOT match basic-easy")
	}

	request = common.Request{
		"path":    "/special/resources/123",
		"headers": common.ConstructMockRequestWithHeaders(map[string][]string{}).Header,
	}
	if limit.match(request) {
		t.Error("Request without Auth header should NOT match basic-easy")
	}
}

// Make sure limit.Add adds requests to different buckets
func TestLimitAdd(t *testing.T) {
	config, err := NewConfiguration("./example.yaml")
	if err != nil {
		t.Error("could not load example configuration")
	}

	limitconfig := limitConfig{
		Interval: 100,
		Max:      3,
		// matches name: Authorization, match: bearer (from example.yaml)
		Matches: config.Limits["basic-simple"].Matches,
		// excludes path: /special/resoures/.*
		Excludes: config.Limits["basic-simple"].Excludes,
		Keys:     config.Limits["basic-simple"].Keys,
	}

	limit, err := newLimit("test-limit", limitconfig, memory.New())
	if err != nil {
		t.Error("Could not initialize test-limit")
	}

	request := common.Request{
		"path": "/special/resources/123",
		"headers": common.ConstructMockRequestWithHeaders(map[string][]string{
			"Authorization": []string{"Basic 12345"},
		}).Header,
	}
	for i := uint(1); i < 4; i++ {
		bucketStatus, err := limit.add(request)
		if err != nil {
			t.Errorf("Error while adding to limit test-limit: %s", err.Error())
		}
		if bucketStatus.Remaining != limitconfig.Max-i {
			t.Errorf("Expected remaining %d, found: %d",
				limitconfig.Max-i, bucketStatus.Remaining)
		}
	}

	bucketStatus, err := limit.add(request)
	if err == nil || err != leakybucket.ErrorFull {
		t.Errorf("Expected leakybucket.ErrorFull error, got: %s", err.Error())
	}

	request2 := common.Request{
		"path": "/special/resources/123",
		"headers": common.ConstructMockRequestWithHeaders(map[string][]string{
			"Authorization": []string{"Basic ABC"},
		}).Header,
	}
	bucketStatus, err = limit.add(request2)
	if err != nil {
		t.Errorf("Error while adding to limit test-limit: %s", err.Error())
	}
	if bucketStatus.Remaining != 2 {
		t.Errorf("Expected remaining %d, found: %d",
			2, bucketStatus.Remaining)
	}
}

package limit

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Clever/leakybucket"
	"github.com/Clever/leakybucket/memory"
	"github.com/Clever/sphinx/common"
	"github.com/Clever/sphinx/config"
	"github.com/Clever/sphinx/matchers"
)

type Storage struct{}

func (s Storage) Create(name string, capacity uint, rate time.Duration) (leakybucket.Bucket, error) {
	return nil, nil
}

// test that matcher errors are bubbled up
func TestBadConfiguration(t *testing.T) {

	configBuf := bytes.NewBufferString(`
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
	config, err := config.LoadAndValidateYaml(configBuf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	if _, err := New("test", config.Limits["test"], Storage{}); err == nil {
		t.Fatal("expected error")
	} else if !strings.Contains(err.Error(), "InvalidMatcherConfig: headers") {
		t.Errorf("Expected a InvalidMatcherConfig error, got different error: %s", err.Error())
	}

}

// assert that the right bucket keys are generated for requests
func TestLimitKeys(t *testing.T) {
	keys, err := resolveLimitKeys(map[string]interface{}{
		"headers": map[string]interface{}{"names": []string{"Authorization", "X-Forwarded-For"}},
		"ip":      []string{""},
	})
	if err != nil {
		t.Fatal(err)
	}
	lim := limit{
		name: "test-limit",
		keys: keys,
	}

	request := common.Request{
		"path": "/resources/123",
		"headers": http.Header{
			"Authorization": []string{"Basic 12345"},
		},
	}
	if lim.bucketName(request) != "test-limit-Authorization:Basic 12345" {
		t.Fatalf("Invalid bucketname for test-limit: %s", lim.bucketName(request))
	}
}

// limit.bucketName creates compound keys from multiple limitkeys
func TestCompoundLimitKeys(t *testing.T) {
	keys, err := resolveLimitKeys(map[string]interface{}{
		"headers": map[string]interface{}{"names": []string{"Authorization", "X-Forwarded-For"}},
		"ip":      []string{""},
	})
	if err != nil {
		t.Fatal(err)
	}
	lim := limit{
		name: "test-limit",
		keys: keys,
	}

	request := common.Request{
		"path":       "/resources/123",
		"remoteaddr": "127.0.0.1",
		"headers": http.Header{
			"Authorization":   []string{"Basic 12345"},
			"X-Forwarded-For": []string{"192.0.0.1"},
		},
	}
	if lim.bucketName(request) !=
		"test-limit-Authorization:Basic 12345-X-Forwarded-For:192.0.0.1-ip:127.0.0.1" {
		t.Fatalf("Invalid compound bucketname for test-limit: %s",
			lim.bucketName(request))
	}
}

// limit.BucketName works when headers are empty
func TestLimitKeyWithEmptyHeaders(t *testing.T) {
	keys, err := resolveLimitKeys(map[string]interface{}{
		"headers": []string{"Authorization", "X-Forwarded-For"},
		"ip":      []string{""},
	})
	if err != nil {
		t.Fatalf("Error while creating limitkeys for test: %s", err)
	}
	lim := limit{
		name: "test-limit",
		keys: keys,
	}

	request := common.Request{
		"path":       "/resources/123",
		"remoteaddr": "127.0.0.1",
		"headers":    http.Header{},
	}
	if lim.bucketName(request) !=
		"test-limit-ip:127.0.0.1" {
		t.Fatalf("Invalid bucketname with no headers for test-limit: %s",
			lim.bucketName(request))
	}
	request = common.Request{
		"path":    "/resources/123",
		"headers": http.Header{},
	}
	if lim.bucketName(request) !=
		"test-limit-" {
		t.Fatalf("Invalid bucketname with no valid request data for test-limit: %s",
			lim.bucketName(request))
	}
}

// limit.bucketName returns consistent names irrespective of header ordering
func TestLimitKeyForConsistentNamingHeaders(t *testing.T) {
	keys, err := resolveLimitKeys(map[string]interface{}{
		"headers": map[string]interface{}{"names": []string{"X-Forwarded-For", "Authorization"}},
		"ip":      []string{""},
	})
	if err != nil {
		t.Fatalf("Error while creating limitkeys for test: %s", err)
	}
	lim := limit{
		name: "test-limit",
		keys: keys,
	}

	requestOne := common.Request{
		"path":       "/resources/123",
		"remoteaddr": "127.0.0.1",
		"headers": http.Header{
			"Authorization":   []string{"Basic 12345"},
			"X-Forwarded-For": []string{"192.0.0.1"},
		},
	}
	requestTwo := common.Request{
		"path": "/resources/123",
		"headers": http.Header{
			"X-Forwarded-For": []string{"192.0.0.1"},
			"Authorization":   []string{"Basic 12345"},
		},
		"remoteaddr": "127.0.0.1",
	}

	if lim.bucketName(requestOne) != lim.bucketName(requestTwo) {
		t.Fatalf("bucketNames do not match with different header order. One: %s.Two: %s",
			lim.bucketName(requestOne), lim.bucketName(requestTwo))
	}

}

// limit.bucketName returns consistent names irrespective of config ordering
func TestLimitKeyForConsistentNamingConfig(t *testing.T) {
	keysOne, err := resolveLimitKeys(map[string]interface{}{
		"headers": map[string]interface{}{"names": []string{"Authorization", "X-Forwarded-For"}},
		"ip":      []string{""},
	})
	if err != nil {
		t.Fatalf("Error while creating limitkeys for test: %s", err)
	}
	limOne := limit{
		name: "test-limit",
		keys: keysOne,
	}

	keysTwo, err := resolveLimitKeys(map[string]interface{}{
		"headers": map[string]interface{}{"names": []string{"X-Forwarded-For", "Authorization"}},
		"ip":      []string{""},
	})
	if err != nil {
		t.Fatalf("Error while creating limitkeys for test: %s", err)
	}
	limTwo := limit{
		name: "test-limit",
		keys: keysTwo,
	}

	request := common.Request{
		"path":       "/resources/123",
		"remoteaddr": "127.0.0.1",
		"headers": http.Header{
			"Authorization":   []string{"Basic 12345"},
			"X-Forwarded-For": []string{"192.0.0.1"},
		},
	}

	if limOne.bucketName(request) != limTwo.bucketName(request) {
		t.Fatalf("bucketNames do not match with different HeaderMatcher config orderging."+
			"One: %s. Two: %s", limOne.bucketName(request), limTwo.bucketName(request))
	}
}

// ensures that Limit.Match exhibits expected behavior
func TestLimitMatch(t *testing.T) {
	data, err := ioutil.ReadFile("../example.yaml")
	if err != nil {
		t.Fatal(err)
	}
	config, err := config.LoadAndValidateYaml(data)
	if err != nil {
		t.Fatal(err)
	}

	// matches name: Authorization, match: bearer (from example.yaml)
	matchers, err := resolveMatchers(config.Limits["basic-simple"].Matches)
	// excludes path: /special/resoures/.*
	excludes, err := resolveMatchers(config.Limits["basic-simple"].Excludes)

	limit := limit{
		name: "test-limit",
		matcher: requestMatcher{
			Matches:  matchers,
			Excludes: excludes,
		},
	}

	request := common.Request{
		"path": "/resources/123",
		"headers": http.Header{
			"Authorization": []string{"Basic 12345"},
		},
	}
	if !limit.Match(request) {
		t.Error("Expected basic-easy to match request")
	}

	request = common.Request{
		"path": "/special/resources/123",
		"headers": http.Header{
			"Authorization": []string{"Basic 12345"},
		},
	}
	if limit.Match(request) {
		t.Error("Request with Excludes path should NOT match basic-easy")
	}

	request = common.Request{
		"path":    "/special/resources/123",
		"headers": http.Header{},
	}
	if limit.Match(request) {
		t.Error("Request without Auth header should NOT match basic-easy")
	}
}

// Make sure limit.Add adds requests to different buckets
func TestLimitAdd(t *testing.T) {
	data, err := ioutil.ReadFile("../example.yaml")
	if err != nil {
		t.Fatal(err)
	}
	conf, err := config.LoadAndValidateYaml(data)
	if err != nil {
		t.Fatal(err)
	}

	limitconfig := config.Limit{
		Interval: 100,
		Max:      3,
		// matches name: Authorization, match: bearer (from example.yaml)
		Matches: conf.Limits["basic-simple"].Matches,
		// excludes path: /special/resoures/.*
		Excludes: conf.Limits["basic-simple"].Excludes,
		Keys:     conf.Limits["basic-simple"].Keys,
	}

	lim, err := New("test-limit", limitconfig, memory.New())
	limit := lim.(*limit)
	if err != nil {
		t.Error("Could not initialize test-limit")
	}

	request := common.Request{
		"path": "/special/resources/123",
		"headers": http.Header{
			"Authorization": []string{"Basic 12345"},
		},
	}
	for i := uint(1); i < 4; i++ {
		bucketStatus, err := limit.Add(request)
		if err != nil {
			t.Errorf("Error while adding to limit test-limit: %s", err.Error())
		}
		if bucketStatus.Remaining != limit.max-i {
			t.Errorf("Expected remaining %d, found: %d",
				limit.max-i, bucketStatus.Remaining)
		}
	}

	bucketStatus, err := limit.Add(request)
	if err == nil {
		t.Fatal("expected error")
	}
	if err != leakybucket.ErrorFull {
		t.Errorf("Expected leakybucket.ErrorFull error, got: %s", err.Error())
	}

	request2 := common.Request{
		"path": "/special/resources/123",
		"headers": http.Header{
			"Authorization": []string{"Basic ABC"},
		},
	}
	bucketStatus, err = limit.Add(request2)
	if err != nil {
		t.Errorf("Error while adding to limit test-limit: %s", err.Error())
	}
	if bucketStatus.Remaining != 2 {
		t.Errorf("Expected remaining %d, found: %d",
			2, bucketStatus.Remaining)
	}
}

type NeverMatch struct{}

func (m NeverMatch) Match(req common.Request) bool {
	return false
}

func createLimit(numMatchers int) Limit {
	neverMatchers := []matchers.Matcher{}
	for i := 0; i < numMatchers; i++ {
		neverMatchers = append(neverMatchers, NeverMatch{})
	}
	limit := &limit{
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
		limit.Match(request)
	}
}

func BenchmarkMatch1(b *testing.B) {
	benchMatch(b, 1)
}

func BenchmarkMatch100(b *testing.B) {
	benchMatch(b, 100)
}

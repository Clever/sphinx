package sphinx

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Clever/sphinx/common"
	"strings"
	"testing"
)

func checkStatusForRequests(ratelimiter RateLimiter,
	request common.Request, num int, expected_statuses []Status) error {

	var statuses []Status
	var err error
	for i := 0; i < num; i++ {
		statuses, err = ratelimiter.Add(request)
		if err != nil {
			return err
		}
	}

	if len(statuses) != len(expected_statuses) {
		return errors.New(fmt.Sprintf("expected to match %d buckets. Got: %d",
			len(expected_statuses), len(statuses)))
	}
	for i, status := range expected_statuses {
		if status.Remaining != statuses[i].Remaining && status.Name != statuses[i].Name {
			return errors.New("Expected 5 requests for the 'bearer/events' limits")
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
		"path": "/v1.1/events/students/123",
		"headers": common.ConstructMockRequestWithHeaders(map[string][]string{
			"Authorization":   []string{"Bearer 12345"},
			"X-Forwarded-For": []string{"IP1", "IP2"},
		}).Header,
		"remoteaddr": "127.0.0.1",
	}
	if err = checkStatusForRequests(
		ratelimiter, request, 5, []Status{
			Status{Remaining: 195, Name: "bearer/events"}}); err != nil {
		t.Error(err)
	}

	request = common.Request{
		"path": "/v1.1/students/123",
		"headers": common.ConstructMockRequestWithHeaders(map[string][]string{
			"Authorization": []string{"Basic 12345"},
		}).Header,
	}

	if err = checkStatusForRequests(
		ratelimiter, request, 1, []Status{
			Status{Remaining: 195, Name: "basic/main"}}); err != nil {
		t.Error(err)
	}
}

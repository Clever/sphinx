package sphinx

import (
	"fmt"
	"github.com/Clever/sphinx/common"
	"testing"
)

// ratelimiter is initialized properly based on config
func TestNewRateLimiter(t *testing.T) {

	config, err := NewConfiguration("./example.yaml")
	if err != nil {
		t.Error("could not load example configuration")
	}

	ratelimiter, err := NewRateLimiter(config)
	if err != nil {
		t.Error(fmt.Sprintf("Error while instantiating ratelimiter: %s", err.Error()))
	}
	if len(ratelimiter.Configuration().Limits) !=
		len(ratelimiter.Limits()) {
		t.Error("expected number of limits in configuration to match instantiated limits")
	}
}

func TestBadConfiguration(t *testing.T) {
}

func TestAdd(t *testing.T) {
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

	for i := 0; i < 4; i++ {
		_, err = ratelimiter.Add(request)
		if err != nil {
			t.Error("Error while adding request to ratelimiter")
		}
	}
	statuses, err := ratelimiter.Add(request)
	if err != nil {
		t.Error("Error while adding request to ratelimiter")
	}
	if len(statuses) != 1 {
		t.Error("expected request to match just one bucket")
	}
	for _, status := range statuses {
		if status.Remaining != 195 && status.Name != "bearer/events" {
			t.Error("Expected 5 requests for the 'bearer/events' limits")
		}
	}
}

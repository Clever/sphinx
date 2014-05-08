package sphinx

import (
	"fmt"
	"github.com/Clever/leakybucket"
	"github.com/Clever/sphinx/common"
	"github.com/Clever/sphinx/config"
	"time"
)

// Status contains the status of a limit.
type Status struct {
	Capacity  uint
	Reset     time.Time
	Remaining uint
	Name      string
}

// NewStatus returns the status of a limit.
func NewStatus(name string, bucket leakybucket.BucketState) Status {

	status := Status{
		Name:      name,
		Capacity:  bucket.Capacity,
		Reset:     bucket.Reset,
		Remaining: bucket.Remaining,
	}

	return status
}

// RateLimiter rate limits requests based on given configuration and limits.
type RateLimiter interface {
	Add(request common.Request) ([]Status, error)
	Configuration() config.Configuration
	Limits() []config.Limit
}

type rateLimiter struct {
	config config.Configuration
	limits []config.Limit
}

func (r *rateLimiter) Limits() []config.Limit {
	return r.limits
}

func (r *rateLimiter) Configuration() config.Configuration {
	return r.config
}

func (r *rateLimiter) Add(request common.Request) ([]Status, error) {
	status := []Status{}
	for _, limit := range r.Limits() {
		if !limit.Match(request) {
			continue
		}
		bucketstate, err := limit.Add(request)
		if err != nil {
			return status, fmt.Errorf("error while adding to Limit: %s. %s",
				limit.Name(), err.Error())
		}
		status = append(status, NewStatus(limit.Name(), bucketstate))
	}
	return status, nil
}

// NewRateLimiter returns a new RateLimiter based on the given configuration.
func NewRateLimiter(config config.Configuration) (RateLimiter, error) {

	rateLimiter := &rateLimiter{config: config, limits: config.Limits()}
	return rateLimiter, nil
}

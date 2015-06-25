package ratelimiter

import (
	"errors"
	"fmt"
	"github.com/Clever/leakybucket"
	leakybucketMemory "github.com/Clever/leakybucket/memory"
	leakybucketRedis "github.com/Clever/leakybucket/redis"
	"github.com/Clever/sphinx/common"
	"github.com/Clever/sphinx/config"
	"github.com/Clever/sphinx/limit"
	"time"
)

// Status contains the status of a limit.
type Status struct {
	Capacity  uint
	Reset     time.Time
	Remaining uint
	Name      string
}

func newStatus(name string, bucket leakybucket.BucketState) Status {

	status := Status{
		Name:      name,
		Capacity:  bucket.Capacity,
		Reset:     bucket.Reset,
		Remaining: bucket.Remaining,
	}

	return status
}

func resolveBucketStore(config map[string]string) (leakybucket.Storage, error) {

	switch config["type"] {
	default:
		return nil, errors.New("must specify one of 'redis' or 'memory' storage")
	case "memory":
		return leakybucketMemory.New(), nil
	case "redis":
		return leakybucketRedis.New("tcp", fmt.Sprintf("%s:%s",
			config["host"], config["port"]))
	}
}

// RateLimiter rate limits requests based on given configuration and limits.
type RateLimiter interface {
	Add(request common.Request) ([]Status, error)
}

type rateLimiter struct {
	limits []limit.Limit
}

func (r *rateLimiter) Add(request common.Request) ([]Status, error) {
	status := []Status{}
	for _, limit := range r.limits {
		if !limit.Match(request) {
			continue
		}
		bucketstate, err := limit.Add(request)
		// Always add the status, so that if we're ratelimited we stil have limit info
		status = append(status, newStatus(limit.Name(), bucketstate))
		if err != nil {
			return status, err
		}
	}
	return status, nil
}

// New returns a new RateLimiter based on the given configuration.
func New(config config.Config) (RateLimiter, error) {

	storage, err := resolveBucketStore(config.Storage)
	if err != nil {
		return nil, err
	}

	limits := []limit.Limit{}
	for name, config := range config.Limits {
		limit, err := limit.New(name, config, storage)
		if err != nil {
			return nil, err
		}
		limits = append(limits, limit)
	}

	rateLimiter := &rateLimiter{limits: limits}
	return rateLimiter, nil
}

// NilStatus for when acting as passive proxy
var NilStatus = Status{
	Capacity:  1,
	Reset:     time.Now(),
	Remaining: 1,
	Name:      "Unknown",
}

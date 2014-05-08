package sphinx

import (
	"errors"
	"fmt"
	"github.com/Clever/leakybucket"
	leakybucketMemory "github.com/Clever/leakybucket/memory"
	leakybucketRedis "github.com/Clever/leakybucket/redis"
	"github.com/Clever/sphinx/common"
	"github.com/Clever/sphinx/limitkeys"
	"github.com/Clever/sphinx/matchers"
	"log"
	"strings"
	"time"
)

func leakyBucketStore(config map[string]string) (leakybucket.Storage, error) {

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

type requestMatcher struct {
	Matches  []matchers.Matcher
	Excludes []matchers.Matcher
}

// Limit stores the information on one of the rate limits being enforced by Sphinx
type Limit struct {
	Name string

	bucketStore leakybucket.Storage
	config      limitConfig
	matcher     requestMatcher
	keys        []limitkeys.LimitKey
}

func (l *Limit) bucketName(request common.Request) string {

	var keyNames []string
	for _, key := range l.keys {
		keyString, err := key.Key(request)

		if err != nil {
			if _, ok := err.(limitkeys.EmptyKeyError); !ok {
				log.Printf("ERROR: Unhandled error while evaluating %s for limit %s. Error: %s",
					key.Type(), l.Name, err.Error())
			}
			// EmptyKeyError is expected for certain requests that do not
			// contain all headerkeys defined in the configuration.
			log.Printf("INFO: %s, %s", l.Name, err.Error())
			continue
		}
		keyNames = append(keyNames, keyString)
	}
	return fmt.Sprintf("%s-%s", l.Name, strings.Join(keyNames, "-"))
}

func (l *Limit) expiry() time.Duration {
	return time.Duration(l.config.Interval) * time.Second
}

func (l *Limit) match(request common.Request) bool {

	// Request does NOT apply if any matcher in Excludes returns true
	for _, matcher := range l.matcher.Excludes {
		match := matcher.Match(request)
		if match {
			return false
		}
	}

	// All Matchers should return true
	for _, matcher := range l.matcher.Matches {
		match := matcher.Match(request)
		if !match {
			return false
		}
	}
	// All matchers returned true
	return true
}

func (l *Limit) add(request common.Request) (leakybucket.BucketState, error) {

	bucket, err := l.bucketStore.Create(l.bucketName(request),
		l.config.Max, l.expiry())
	if err != nil {
		return leakybucket.BucketState{}, err
	}

	return bucket.Add(1)
}

func newLimit(name string, config limitConfig, storage leakybucket.Storage) (*Limit, error) {

	limit := Limit{}
	limit.Name = name
	limit.bucketStore = storage
	limit.config = config

	limit.matcher = requestMatcher{}

	matches, err := resolveMatchers(config.Matches)
	if err != nil {
		log.Printf("Failed to load matchers for LIMIT:%s, ERROR:%s", name, err)
		return &limit, err
	}
	excludes, err := resolveMatchers(config.Excludes)
	if err != nil {
		log.Printf("Failed to load excludes for LIMIT:%s, ERROR:%s.", name, err)
		return &limit, err
	}
	limitkeys, err := resolveLimitKeys(config.Keys)
	if err != nil {
		log.Printf("Failed to load keys for LIMIT:%s, ERROR:%s.", name, err)
		return &limit, err
	}

	limit.matcher.Matches = matches
	limit.matcher.Excludes = excludes
	limit.keys = limitkeys
	return &limit, nil
}

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
	Configuration() Configuration
	Limits() []*Limit
	SetLimits([]*Limit)
}

type sphinxRateLimiter struct {
	configuration Configuration
	limits        []*Limit
}

func (r *sphinxRateLimiter) Limits() []*Limit {
	return r.limits
}

func (r *sphinxRateLimiter) Configuration() Configuration {
	return r.configuration
}

func (r *sphinxRateLimiter) SetLimits(limits []*Limit) {
	r.limits = limits
}

func (r *sphinxRateLimiter) Add(request common.Request) ([]Status, error) {
	var status []Status
	for _, limit := range r.Limits() {
		if match := limit.match(request); match {
			bucketstate, err := limit.add(request)
			if err != nil {
				return status,
					fmt.Errorf("error while adding to Limit: %s. %s",
						limit.Name, err.Error())
			}
			status = append(status, NewStatus(limit.Name, bucketstate))
		}
	}
	return status, nil
}

// NewRateLimiter returns a new RateLimiter based on the given configuration.
func NewRateLimiter(config Configuration) (RateLimiter, error) {

	rateLimiter := sphinxRateLimiter{
		configuration: config,
	}
	storage, err := leakyBucketStore(config.Storage)
	if err != nil {
		return &rateLimiter, err
	}

	var limits []*Limit
	for name, config := range config.Limits {
		limit, err := newLimit(name, config, storage)
		if err != nil {
			return &rateLimiter, err
		}
		limits = append(limits, limit)
	}
	rateLimiter.SetLimits(limits)

	return &rateLimiter, nil
}

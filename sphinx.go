package sphinx

import (
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

func LeakyBucketStore(config map[string]string) (leakybucket.Storage, error) {

	switch config["type"] {
	// default is leakybucket memory
	default:
		return leakybucketMemory.New(), nil
	case "redis":
		redisstore, err := leakybucketRedis.New("tcp", fmt.Sprintf("%s:%s",
			config["host"], config["port"]))

		return redisstore, err
	}
}

type RequestMatcher struct {
	Matches  []matchers.Matcher
	Excludes []matchers.Matcher
}

type Limit struct {
	Name string

	bucketStore leakybucket.Storage
	config      LimitConfig
	matcher     RequestMatcher
	keys        []limitkeys.LimitKey
}

func (l *Limit) BucketName(request common.Request) string {

	var keyNames []string
	for _, key := range l.keys {
		keyString, err := key.Key(request)
		if err != nil {
			continue
		}
		keyNames = append(keyNames, keyString)
	}

	return fmt.Sprintf("%s-%s", l.Name, strings.Join(keyNames, "-"))
}

func (l *Limit) Match(request common.Request) bool {

	// Request does NOT apply if any matcher in Excludes returns true
	for _, matcher := range l.matcher.Excludes {
		match := matcher.Match(request)
		if match {
			return false
		}
	}

	// At least one matcher in Matches should return true
	for _, matcher := range l.matcher.Matches {
		match := matcher.Match(request)
		if match {
			return true
		}
	}

	// does not apply to any matcher in this limit
	return false
}

func (l *Limit) Add(request common.Request) (leakybucket.BucketState, error) {

	var bucketstate leakybucket.BucketState
	expiry := time.Duration(l.config.Interval) * time.Second
	bucket, err := l.bucketStore.Create(l.BucketName(request),
		l.config.Max, expiry)
	if err != nil {
		return bucketstate, err
	}

	bucketstate, err = bucket.Add(1)
	if err != nil {
		return bucketstate, err
	}

	return bucketstate, nil
}

func NewLimit(name string, config LimitConfig, storage leakybucket.Storage) *Limit {

	limit := Limit{}
	limit.Name = name
	limit.bucketStore = storage
	limit.config = config

	limit.matcher = RequestMatcher{}
	var err error
	limit.matcher.Matches, err = ResolveMatchers(config.Matches)
	limit.matcher.Excludes, err = ResolveMatchers(config.Excludes)
	if err != nil {
		log.Fatalf("Failed to load matchers.", err)
	}

	return &limit
}

type Status struct {
	Capacity  uint
	Reset     time.Time
	Remaining uint
	Name      string
}

func NewStatus(name string, bucket leakybucket.BucketState) Status {

	status := Status{
		Name:      name,
		Capacity:  bucket.Capacity,
		Reset:     bucket.Reset,
		Remaining: bucket.Remaining,
	}

	return status
}

type RateLimiter interface {
	Add(request common.Request) ([]Status, error)
	Configuration() Configuration
	Limits() []*Limit
	SetLimits([]*Limit)
}

type SphinxRateLimiter struct {
	configuration Configuration
	limits        []*Limit
}

func (r *SphinxRateLimiter) Limits() []*Limit {
	return r.limits
}

func (r *SphinxRateLimiter) Configuration() Configuration {
	return r.configuration
}

func (r *SphinxRateLimiter) SetLimits(limits []*Limit) {
	r.limits = limits
}

func (r *SphinxRateLimiter) Add(request common.Request) ([]Status, error) {
	var status []Status
	for _, limit := range r.Limits() {
		if match := limit.Match(request); match {
			bucketstate, err := limit.Add(request)
			if err != nil {
				return status,
					fmt.Errorf("Error while adding to Limit: %s. %s",
						limit.Name, err.Error())
			}
			status = append(status, NewStatus(limit.Name, bucketstate))
		}
	}
	return status, nil
}

func NewRateLimiter(config Configuration) (RateLimiter, error) {

	rateLimiter := SphinxRateLimiter{
		configuration: config,
	}
	storage, err := LeakyBucketStore(config.Storage)
	if err != nil {
		return &rateLimiter, err
	}

	var limits []*Limit
	for name, config := range config.Limits {
		limits = append(limits, NewLimit(name, config, storage))
	}
	rateLimiter.SetLimits(limits)

	return &rateLimiter, nil
}

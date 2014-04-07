package sphinx

import (
	"fmt"
	"github.com/Clever/leakybucket"
	leakybucketMemory "github.com/Clever/leakybucket/memory"
	"github.com/Clever/sphinx/common"
	"github.com/Clever/sphinx/limitkeys"
	"github.com/Clever/sphinx/matchers"
	"log"
	"strings"
	"time"
)

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
		keyNames = append(keyNames, key.GetKey(request))
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
		if !match {
			return true
		}
	}

	// does not apply to any matcher in this limit
	return false
}

func (l *Limit) Add(request common.Request) (leakybucket.BucketState, error) {

	var bucketstate leakybucket.BucketState
	bucket, err := l.bucketStore.Create(l.BucketName(request),
		l.config.Max, l.config.Interval)

	if err != nil {
		return bucketstate, err
	}
	bucketstate, err = bucket.Add(1)
	if err != nil {
		return bucketstate, err
	}

	return bucketstate, nil
}

func NewLimit(name string, config LimitConfig) Limit {

	limit := Limit{}
	limit.Name = name
	limit.bucketStore = leakybucketMemory.New()
	limit.config = config

	limit.matcher = RequestMatcher{}
	var err error
	limit.matcher.Matches, err = ResolveMatchers(config.Matches)
	limit.matcher.Excludes, err = ResolveMatchers(config.Excludes)
	if err != nil {
		log.Panicf("Failed to load matchers.", err)
	}

	return limit
}

type Status struct {
	Capacity  uint
	Reset     time.Time
	Remaining uint
	Name      string
}

func NewStatus(name string, bucket leakybucket.BucketState) Status {

	status := Status{}
	status.Name = name
	status.Capacity = bucket.Capacity
	status.Reset = bucket.Reset
	status.Remaining = bucket.Remaining

	return status
}

type RateLimiter interface {
	Configuration() Configuration
	Limits() []Limit
	SetLimits([]Limit)
	Add(request common.Request) ([]Status, error)
}

type SphinxRateLimiter struct {
	configuration Configuration
	limits        []Limit
}

func (r *SphinxRateLimiter) Limits() []Limit {
	return r.limits
}

func (r *SphinxRateLimiter) Configuration() Configuration {
	return r.configuration
}

func (r *SphinxRateLimiter) SetLimits(limits []Limit) {
	r.limits = limits
}

func (r *SphinxRateLimiter) Add(request common.Request) ([]Status, error) {
	var status []Status
	for _, limit := range r.Limits() {
		if match := limit.Match(request); match {
			bucketstate, err := limit.Add(request)
			if err == nil {
				//TODO SOMETHING
			}
			status = append(status, NewStatus(limit.Name, bucketstate))
		}
	}
	return status, nil
}

func NewDaemon(config Configuration) {

	rateLimiter := SphinxRateLimiter{}

	limits := make([]Limit, len(config.Limits))
	for name, config := range config.Limits {
		limits = append(limits, NewLimit(name, config))
	}
	rateLimiter.SetLimits(limits)
}

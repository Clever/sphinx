package sphinx

import (
	"github.com/Clever/leakybucket"
	leakybucketMemory "github.com/Clever/leakybucket/memory"
	"github.com/Clever/sphinx/common"
	"github.com/Clever/sphinx/limitkeys"
	"github.com/Clever/sphinx/matchers"
	"log"
	"time"
)

type RequestMatcher struct {
	Matches  []matchers.Matcher
	Excludes []matchers.Matcher
}

type Limit struct {
	Name string

	bucketStore leakybucket.Storage
	buckets     map[string]leakybucket.Bucket
	config      LimitConfig
	matcher     RequestMatcher
	keys        []limitkeys.LimitKey
}

func (l *Limit) getBucketName(request common.Request) string {

	// compute bucketName = Limit.Name + concat(keys)
	// eg. bearer/events-header:authentication:ABCD-request:ip:172.0.0.1

	return "nothing"
}

func (l *Limit) Match(request common.Request) (bool, error) {
	// match with matches and excludes
	return false, nil
}

func (l *Limit) Add(request common.Request) (leakybucket.Bucket, error) {
	// getBucketName
	// if bucket already exists add to bucket
	// if does not exist create new bucket with name
	return l.bucketStore.Create("", 0, 0)
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

func NewStatus(name string, bucket leakybucket.Bucket) Status {

	status := Status{}
	status.Name = name
	status.Capacity = bucket.Capacity()
	status.Reset = bucket.Reset()
	status.Remaining = bucket.Remaining()

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
		if match, _ := limit.Match(request); match {
			bucket, err := limit.Add(request)
			if err == nil {
				//DO SOMETHING
			}
			status = append(status, NewStatus(limit.Name, bucket))
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

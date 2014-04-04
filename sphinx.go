package sphinx

import (
	"github.com/Clever/leakybucket"
	leakybucketMemory "github.com/Clever/leakybucket/memory"
	"github.com/Clever/sphinx/matchers"
	"log"
	"regexp"
	"time"
)

type Request map[string]interface{}

type RequestMatcher struct {
	Matches  map[string][]matchers.Matcher
	Excludes map[string][]matchers.Matcher
}

func (r *RequestMatcher) AddMatches(name string, rules []string) error {

	return nil
}

func (r *RequestMatcher) AddExcludes(name string, rules []string) error {

	return nil
}

type Limit struct {
	Name string

	bucketStore leakybucket.Storage
	buckets     map[string]leakybucket.Bucket
	config      LimitConfig
	matcher     RequestMatcher
}

func (l *Limit) getBucketName(request map[string]string) string {

	// compute bucketName = Limit.Name + concat(keys)
	// eg. bearer/events-header:authentication:ABCD-request:ip:172.0.0.1

	return "nothing"
}

func (l *Limit) Match(request map[string]string) (bool, error) {
	// match with matches and excludes
	return false, nil
}

func (l *Limit) Add(request Request) (leakybucket.Bucket, error) {
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

	return limit

}

type Status struct {
	Capacity  uint
	Reset     time.Time
	Remaining uint
	Name      string
}

type RateLimiter interface {
	Configuration() Configuration
	Limits() []Limit
	SetLimits([]Limit)
	Add(request Request) ([]Status, error)
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

func (r *SphinxRateLimiter) Add(request Request) ([]Status, error) {
	// status := make([]status)
	// for limit in limits
	//   if limit.Match(request)
	//     buckets, err := limit.Add(request)
	//       for bucket in buckets
	//         status = append(status, NewStatus)
	// return status, nil
	return nil, nil
}

func NewDaemon(config Configuration) {

	rateLimiter := SphinxRateLimiter{}

	limits := make([]Limit, len(config.Limits))
	for name, config := range config.Limits {
		limits = append(limits, NewLimit(name, config))
	}
	rateLimiter.SetLimits(limits)
}

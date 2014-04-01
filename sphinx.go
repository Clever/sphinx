package main

import (
	leakybucket "github.com/Clever/leakybucket"
	leakybucketMemory "github.com/Clever/leakybucket/memory"
	//"log"
	"regexp"
)

type RequestMatcher struct {
	Matches  map[string][]*regexp.Regexp
	Excludes map[string][]*regexp.Regexp
}

func compileRules(rules []string) ([]*regexp.Regexp, error) {
	regexps := []*regexp.Regexp{}

	for _, rule := range rules {
		compiled, err := regexp.Compile(rule)
		if err != nil {
			return nil, err
		}
		regexps = append(regexps, compiled)
	}

	return regexps, nil
}

func (r *RequestMatcher) AddMatches(name string, rules []string) error {

	compiledRules, err := compileRules(rules)
	if err != nil {
		return err
	}

	r.Matches[name] = compiledRules
	return nil
}

func (r *RequestMatcher) AddExcludes(name string, rules []string) error {

	compiledRules, err := compileRules(rules)
	if err != nil {
		return err
	}

	r.Excludes[name] = compiledRules
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

func (l *Limit) Add(request string /*sphinx.Request*/) (leakybucket.Bucket, error) {
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

type RateLimiter struct {
	Configuration Configuration
	Limits        []Limit
}

func (r *RateLimiter) Add(request map[string]string) /*(buckets []leakybucket.Bucket, error)*/ {

	// check match for every limit
	// add to all buckets that match
	// return matched buckets
	return
}

func NewSphinxDaemon(config Configuration) {

	rateLimiter := RateLimiter{}

	for name, config := range config.Limits {
		rateLimiter.Limits = append(rateLimiter.Limits, NewLimit(name, config))

	}
}

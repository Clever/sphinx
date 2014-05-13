package limit

import (
	"fmt"
	"github.com/Clever/leakybucket"
	"github.com/Clever/sphinx/common"
	"github.com/Clever/sphinx/config"
	"github.com/Clever/sphinx/limitkeys"
	"github.com/Clever/sphinx/matchers"
	"log"
	"strings"
	"time"
)

// Limit has methods for matching and adding to a limit
type Limit interface {
	Name() string
	Match(common.Request) bool
	Add(common.Request) (leakybucket.BucketState, error)
}

type requestMatcher struct {
	Matches  []matchers.Matcher
	Excludes []matchers.Matcher
}

type limit struct {
	name     string
	storage  leakybucket.Storage
	keys     []limitkeys.LimitKey
	matcher  requestMatcher
	max      uint
	interval uint
}

func (l limit) Name() string {
	return l.name
}

func (l limit) Match(request common.Request) bool {

	// Request does NOT apply if any matcher in Excludes returns true
	for _, matcher := range l.matcher.Excludes {
		if matcher.Match(request) {
			return false
		}
	}

	// All Matchers should return true
	for _, matcher := range l.matcher.Matches {
		if !matcher.Match(request) {
			return false
		}
	}
	// All matchers returned true
	return true
}

func (l limit) Add(request common.Request) (leakybucket.BucketState, error) {

	bucket, err := l.storage.Create(l.bucketName(request),
		l.max, l.expiry())
	if err != nil {
		return leakybucket.BucketState{}, err
	}

	return bucket.Add(1)
}

func (l limit) bucketName(request common.Request) string {

	keyNames := []string{}
	for _, key := range l.keys {
		keyString, err := key.Key(request)

		if err != nil {
			if _, ok := err.(limitkeys.EmptyKeyError); !ok {
				log.Printf("ERROR: Unhandled error while evaluating %s for limit %s. Error: %s",
					key.Type(), l.Name(), err.Error())
			}
			// EmptyKeyError is expected for certain requests that do not
			// contain all headerkeys defined in the configuration.
			log.Printf("INFO: %s, %s", l.Name(), err.Error())
			continue
		}
		keyNames = append(keyNames, keyString)
	}
	return fmt.Sprintf("%s-%s", l.Name(), strings.Join(keyNames, "-"))
}

func (l limit) expiry() time.Duration {
	return time.Duration(l.interval) * time.Second
}

// New creates a new Limit
func New(name string, config config.Limit, storage leakybucket.Storage) (Limit, error) {

	limit := &limit{
		name:     name,
		storage:  storage,
		interval: config.Interval,
		max:      config.Max,
		matcher:  requestMatcher{},
	}

	matches, err := resolveMatchers(config.Matches)
	if err != nil {
		log.Printf("Failed to load matchers for LIMIT:%s, ERROR:%s", name, err)
		return nil, err
	}
	limit.matcher.Matches = matches
	excludes, err := resolveMatchers(config.Excludes)
	if err != nil {
		log.Printf("Failed to load excludes for LIMIT:%s, ERROR:%s.", name, err)
		return nil, err
	}
	limit.matcher.Excludes = excludes

	limitkeys, err := resolveLimitKeys(config.Keys)
	if err != nil {
		log.Printf("Failed to load keys for LIMIT:%s, ERROR:%s.", name, err)
		return nil, err
	}

	limit.keys = limitkeys
	return limit, nil
}

func resolveMatchers(matchersConfig map[string]interface{}) ([]matchers.Matcher, error) {

	resolvedMatchers := []matchers.Matcher{}

	// try and setup Matches to the actual config object defined by matchers
	for key, config := range matchersConfig {
		factory := matchers.MatcherFactoryFinder(key)
		if factory == nil {
			return resolvedMatchers, fmt.Errorf("could not find matcher for %s", key)
		}
		matcher, err := factory.Create(config)
		if err != nil {
			return resolvedMatchers, err
		}
		resolvedMatchers = append(resolvedMatchers, matcher)
	}
	return resolvedMatchers, nil
}

func resolveLimitKeys(limitkeysConfig map[string]interface{}) ([]limitkeys.LimitKey, error) {

	resolvedLimitkeys := []limitkeys.LimitKey{}

	for name, config := range limitkeysConfig {
		switch name {
		case "headers":
			keys, err := limitkeys.NewHeaderLimitKeys(config)
			if err != nil {
				return []limitkeys.LimitKey{}, err
			}
			resolvedLimitkeys = append(resolvedLimitkeys, keys...)
		case "ip":
			keys, err := limitkeys.NewIPLimitKeys(config)
			if err != nil {
				return []limitkeys.LimitKey{}, err
			}
			resolvedLimitkeys = append(resolvedLimitkeys, keys...)
		default:
			return []limitkeys.LimitKey{},
				fmt.Errorf("only header and ip limitkeys allowed. Found: %s", name)
		}
	}

	return resolvedLimitkeys, nil
}

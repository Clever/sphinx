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
	"gopkg.in/v1/yaml"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"strings"
	"time"
)

// Configuration holds the limits and proxy configurations
type Configuration interface {
	Proxy() Proxy
	Limits() []Limit
}

type configuration struct {
	proxy  Proxy
	limits []Limit
}

func (c configuration) Proxy() Proxy {
	return c.proxy
}
func (c configuration) Limits() []Limit {
	return c.limits
}

// Proxy holds the proxy config
type Proxy struct {
	Handler string
	Host    string
	Listen  string
}

// Limit has methods for matching and adding to a limit
type Limit interface {
	Name() string
	Match(common.Request) bool
	Add(common.Request) (leakybucket.BucketState, error)
}

type limit struct {
	name    string
	storage leakybucket.Storage
	config  limitYaml
	keys    []limitkeys.LimitKey
	matcher requestMatcher
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
		l.config.Max, l.expiry())
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
			log.Printf("INFO: %s, %s", l.Name, err.Error())
			continue
		}
		keyNames = append(keyNames, keyString)
	}
	return fmt.Sprintf("%s-%s", l.Name(), strings.Join(keyNames, "-"))
}

func (l limit) expiry() time.Duration {
	return time.Duration(l.config.Interval) * time.Second
}

func newLimit(name string, config limitYaml, storage leakybucket.Storage) (Limit, error) {

	limit := &limit{
		name:    name,
		storage: storage,
		config:  config,
		matcher: requestMatcher{},
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

// Configuration holds current Sphinx configuration
type configYaml struct {
	Proxy   Proxy
	Limits  map[string]limitYaml
	Storage map[string]string
}

type limitYaml struct {
	Interval uint
	Max      uint
	Keys     map[string]interface{}
	Matches  map[string]interface{}
	Excludes map[string]interface{}
}

func loadAndValidateConfig(data []byte) (configYaml, error) {

	config := configYaml{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Print("Failed to parse data in configuration. Aborting")
		return config, err
	}

	if config.Proxy.Handler == "" {
		return config, fmt.Errorf("proxy.handler not set")
	}
	if config.Proxy.Host == "" {
		return config, fmt.Errorf("proxy.host not set")
	}
	if _, _, err := net.SplitHostPort(config.Proxy.Listen); err != nil {
		return config, fmt.Errorf("invalid proxy.listen. Should be like host:port or :port")
	}

	if _, err := url.ParseRequestURI(config.Proxy.Host); err != nil {
		return config,
			errors.New("Could not parse proxy.host. " +
				"Must include scheme (eg. https://example.com)")
	}
	if len(config.Limits) < 1 {
		return config, fmt.Errorf("no limits definied")
	}

	for name, limit := range config.Limits {
		if len(limit.Keys) == 0 {
			return config, fmt.Errorf("must set at least one key for limit: %s", name)
		}
		if limit.Interval < 1 {
			return config, fmt.Errorf("interval must be set > 1 for limit: %s", name)
		}
		if limit.Max < 1 {
			return config, fmt.Errorf("max must be set > 1 for limit: %s", name)
		}
	}

	store, ok := config.Storage["type"]
	if !ok {
		return config, fmt.Errorf("storage type must be set")
	}
	switch strings.ToLower(store) {
	default:
		return config, fmt.Errorf("storage type needs to be memory or redis")
	case "redis":
		if _, ok := config.Storage["host"]; !ok {
			return config, fmt.Errorf("storage host must be set for Redis")
		}
		if _, ok := config.Storage["port"]; !ok {
			return config, fmt.Errorf("storage port must be set for Redis")
		}
	case "memory":
		// nothing to do here
	}

	return config, nil
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

	for name, values := range limitkeysConfig {
		switch name {
		case "headers":
			headernames := []string{}
			common.ReMarshal(values, &headernames)
			for _, headername := range headernames {
				resolvedLimitkeys = append(resolvedLimitkeys,
					limitkeys.NewHeaderLimitKey(headername))
			}
		case "ip":
			resolvedLimitkeys = append(resolvedLimitkeys, limitkeys.NewIPLimitKey())
		default:
			return []limitkeys.LimitKey{},
				fmt.Errorf("only header and ip limitkeys allowed. Found: %s", name)
		}
	}

	return resolvedLimitkeys, nil
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

func parseYaml(config configYaml) (Configuration, error) {
	storage, err := resolveBucketStore(config.Storage)
	if err != nil {
		return nil, err
	}

	limits := []Limit{}
	for name, config := range config.Limits {
		limit, err := newLimit(name, config, storage)
		if err != nil {
			return nil, err
		}
		limits = append(limits, limit)
	}

	return &configuration{
		proxy:  config.Proxy,
		limits: limits,
	}, nil
}

// NewConfiguration takes in a path to a configuration yaml and returns a Configuration.
func NewConfiguration(path string) (Configuration, error) {

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil,
			fmt.Errorf("failed to read %s. Aborting with error: %s", path, err.Error())
	}
	yaml, err := loadAndValidateConfig(data)
	if err != nil {
		return nil, err
	}
	return parseYaml(yaml)
}

package sphinx

import (
	"errors"
	"fmt"
	"github.com/Clever/sphinx/matchers"
	"gopkg.in/v1/yaml"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"strings"
)

// Configuration holds current Sphinx configuration
type Configuration struct {
	Proxy   proxy
	Limits  map[string]limitConfig
	Storage map[string]string
}

type proxy struct {
	Handler string
	Host    string
	Listen  string
}

// LimitConfig contains configuration for a Limit
type limitConfig struct {
	Interval uint
	Max      uint
	Keys     map[string]string
	Matches  map[string]interface{}
	Excludes map[string]interface{}
}

func loadAndValidateConfig(data []byte) (Configuration, error) {

	config := Configuration{}
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

// NewConfiguration takes in a path to a configuration yaml and returns a Configuration.
func NewConfiguration(path string) (Configuration, error) {

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return Configuration{},
			fmt.Errorf("failed to read %s. Aborting with error: %s", path, err.Error())
	}
	return loadAndValidateConfig(data)
}

package sphinx

import (
	"fmt"
	"github.com/Clever/sphinx/matchers"
	"gopkg.in/v1/yaml"
	"io/ioutil"
	"log"
	"net/url"
	"time"
)

type Configuration struct {
	Forward Forward
	Limits  map[string]LimitConfig
	Storage map[string]string
}

type Forward struct {
	Scheme string
	Host   string
	Listen string
}

type LimitConfig struct {
	Interval time.Duration
	Max      uint
	Keys     map[string]string
	Matches  map[string]interface{}
	Excludes map[string]interface{}
}

func panicWithError(err error, message string) {
	if err != nil {
		log.Print(message)
		log.Panic(err)
	}
}

func loadAndValidateConfig(data []byte) (Configuration, error) {

	config := Configuration{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Print("Failed to parse data in configuration. Aborting")
		return config, err
	}

	if config.Forward.Scheme == "" {
		return config, fmt.Errorf("forward.scheme not set")
	}
	if config.Forward.Host == "" {
		return config, fmt.Errorf("forward.host not set")
	}
	if _, err := url.Parse(config.Forward.Host); err != nil {
		return config, fmt.Errorf("could not parse forward.host")
	}
	if len(config.Limits) < 1 {
		return config, fmt.Errorf("No limits definied")
	}

	for name, limit := range config.Limits {
		if limit.Interval < 1 {
			return config, fmt.Errorf("Interval must be set > 1 for limit: %s", name)
		}
		if limit.Max < 1 {
			return config, fmt.Errorf("Max must be set > 1 for limit: %s", name)
		}
	}

	store, ok := config.Storage["type"]
	if !ok {
		return config, fmt.Errorf("leakybucket:store must be set.")
	}
	switch store {
	default:
		return config, fmt.Errorf("Storage type needs to be memory or redis")
	case "redis":
		if _, ok := config.Storage["host"]; !ok {
			config.Storage["host"] = "localhost"
		}
		if _, ok := config.Storage["port"]; !ok {
			config.Storage["port"] = "6379"
		}
	case "memory":
		// nothing to do here
	}

	return config, nil
}

func ResolveMatchers(matchersConfig map[string]interface{}) ([]matchers.Matcher, error) {

	resolvedMatchers := []matchers.Matcher{}

	// try and setup Matches to the actual config object defined by matchers
	for key, config := range matchersConfig {
		factory := matchers.MatcherFactoryFinder(key)
		if factory == nil {
			return resolvedMatchers, fmt.Errorf("Could not find matcher for %s", key)
		}
		matcher, err := factory.Create(config)
		if err != nil {
			return resolvedMatchers, err
		}
		resolvedMatchers = append(resolvedMatchers, matcher)
	}
	return resolvedMatchers, nil
}

func NewConfiguration(path string) (Configuration, error) {
	data, err := ioutil.ReadFile(path)
	return Configuration{},
		fmt.Errorf("Failed to read %s. Aborting with error: %s", path, err.Error())

	config, load_err := loadAndValidateConfig(data)
	return config, load_err
}

package config

import (
	"errors"
	"fmt"
	"gopkg.in/v1/yaml"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"strings"
)

// Config holds the yaml data for the config file
type Config struct {
	Proxy   Proxy
	Limits  map[string]Limit
	Storage map[string]string
}

// Proxy holds the yaml data for the proxy option in the config file
type Proxy struct {
	Handler string
	Host    string
	Listen  string
}

// Limit holds the yaml data for one of the limits in the config file
type Limit struct {
	Interval uint
	Max      uint
	Keys     map[string]interface{}
	Matches  map[string]interface{}
	Excludes map[string]interface{}
}

// LoadYaml loads byte data for a yaml file into a Config
func LoadYaml(data []byte) (Config, error) {
	config := Config{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Print("Failed to parse data in configuration. Aborting")
		return Config{}, err
	}
	return config, nil
}

// ValidateConfig validates that a Config has all the required fields
func ValidateConfig(config Config) error {
	if config.Proxy.Handler == "" {
		return fmt.Errorf("proxy.handler not set")
	}
	if config.Proxy.Host == "" {
		return fmt.Errorf("proxy.host not set")
	}
	if _, _, err := net.SplitHostPort(config.Proxy.Listen); err != nil {
		return fmt.Errorf("invalid proxy.listen. Should be like host:port or :port")
	}

	if _, err := url.ParseRequestURI(config.Proxy.Host); err != nil {
		return errors.New("could not parse proxy.host. Must include scheme (eg. https://example.com)")
	}
	if len(config.Limits) < 1 {
		return fmt.Errorf("no limits definied")
	}

	for name, limit := range config.Limits {
		if len(limit.Keys) == 0 {
			return fmt.Errorf("must set at least one key for limit: %s", name)
		}
		if limit.Interval < 1 {
			return fmt.Errorf("interval must be set > 1 for limit: %s", name)
		}
		if limit.Max < 1 {
			return fmt.Errorf("max must be set > 1 for limit: %s", name)
		}
	}

	store, ok := config.Storage["type"]
	if !ok {
		return fmt.Errorf("storage type must be set")
	}
	switch strings.ToLower(store) {
	default:
		return fmt.Errorf("storage type needs to be memory or redis")
	case "redis":
		if _, ok := config.Storage["host"]; !ok {
			return fmt.Errorf("storage host must be set for Redis")
		}
		if _, ok := config.Storage["port"]; !ok {
			return fmt.Errorf("storage port must be set for Redis")
		}
	case "memory":
		// nothing to do here
	}

	return nil
}

// LoadAndValidateYaml turns a sequence of bytes into a Config and validates that all the necessary
// fields are set
// TODO (z): These should all be private, but right now tests depend on parsing bytes into yaml
func LoadAndValidateYaml(data []byte) (Config, error) {
	config, err := LoadYaml(data)
	if err != nil {
		return Config{}, err
	}
	if err := ValidateConfig(config); err != nil {
		return Config{}, err
	}
	return config, nil
}

// New takes in a path to a configuration yaml and returns a Configuration.
func New(path string) (Config, error) {

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{},
			fmt.Errorf("failed to read %s. Aborting with error: %s", path, err.Error())
	}
	return LoadAndValidateYaml(data)
}

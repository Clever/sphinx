package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"strings"

	"gopkg.in/yaml.v1"
)

// Config holds the yaml data for the config file
type Config struct {
	Proxy       Proxy
	HealthCheck HealthCheck `yaml:"health-check"`
	Limits      map[string]Limit
	Storage     map[string]string
}

// Proxy holds the yaml data for the proxy option in the config file
type Proxy struct {
	Handler      string
	Host         string
	Listen       string
	AllowOnError bool `yaml:"allow-on-error"`
}

// HealthCheck holds the yaml data for how to run the health check service.
type HealthCheck struct {
	Port     string
	Endpoint string
	Enabled  bool
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
// TODO (z): These should all be private, but right now tests depend on parsing bytes into yaml
func LoadYaml(data []byte) (Config, error) {
	config := Config{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Print("Failed to parse data in configuration. Aborting")
		return Config{}, err
	}
	return config, nil
}

// ValidateConfig validates that a Config has all the required fields
// TODO (z): These should all be private, but right now tests depend on parsing bytes into yaml
func ValidateConfig(config Config) error {
	// NOTE: tests depend on the order of these checks.
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

	// HealthCheck section is optional.
	if config.HealthCheck.Enabled {
		colonIdx := strings.LastIndex(config.Proxy.Listen, ":") + 1
		proxyPort := config.Proxy.Listen[colonIdx:]
		if config.HealthCheck.Port == proxyPort {
			return fmt.Errorf("health service port cannot match proxy.listen port")
		}
	}

	if len(config.Limits) < 1 {
		return fmt.Errorf("no limits defined")
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
	case "dynamodb":
		if _, ok := config.Storage["region"]; !ok {
			return fmt.Errorf("storage region must be set for DynamoDB")
		}
		if _, ok := config.Storage["table"]; !ok {
			return fmt.Errorf("storage table must be set for DynamoDB")
		}
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

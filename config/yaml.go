package config

import (
	"errors"
	"fmt"
	"gopkg.in/v1/yaml"
	"log"
	"net"
	"net/url"
	"strings"
)

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

func loadAndValidateYaml(data []byte) (configYaml, error) {

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

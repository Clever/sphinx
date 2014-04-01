package main

import (
	"fmt"
	"gopkg.in/v1/yaml"
	"io/ioutil"
	"log"
	"time"
)

type Configuration struct {
	Forward Forward
	Limits  map[string]LimitConfig
}

type Forward struct {
	Scheme string
	Host   string
}

type LimitConfig struct {
	Interval time.Duration
	Max      uint
	Keys     []string
	Matches  map[string][]string
	Excludes map[string][]string
}

func panicWithError(err error, message string) {
	if err != nil {
		log.Print(message)
		log.Panic(err)
	}
}

func loadAndValidateConfig(data []byte) (Configuration, error) {

	config := Configuration{}
	err := yaml.Unmarshal(data, &config)
	panicWithError(err, fmt.Sprintf("Failed to parse data in configuration. Aborting"))

	if config.Forward.Scheme == "" {
		return config, fmt.Errorf("forward.scheme not set")
	}
	if config.Forward.Host == "" {
		return config, fmt.Errorf("forward.host not set")
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

	return config, nil
}

func NewConfiguration(path string) Configuration {
	data, err := ioutil.ReadFile(path)
	panicWithError(err, fmt.Sprintf("Failed to read configuration %s. Aborting", path))

	config, load_err := loadAndValidateConfig(data)
	panicWithError(load_err, "Failed to validate configuration")

	return config
}

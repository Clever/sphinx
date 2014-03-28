package main

import (
	"fmt"
	"gopkg.in/v1/yaml"
	"io/ioutil"
	"log"
)

type Configuration struct {
	Forward Forward
	Buckets map[string]BucketConfig
}

type Forward struct {
	Scheme string
	Host   string
}

type BucketConfig struct {
	Interval int
	Limit    int
	Keys     []string
	Matches  Rules
	Excludes Rules
}

type Rules struct {
	Headers []string
	Paths   []string
}

func panicWithError(err error, message string) {
	if err != nil {
		log.Panic(message, err)
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
	if len(config.Buckets) < 1 {
		return config, fmt.Errorf("No buckets definied")
	}

	for name, bucket := range config.Buckets {
		if bucket.Interval < 1 {
			return config, fmt.Errorf("Interval must be set > 1 for bucket: %s", name)
		}
		if bucket.Limit < 1 {
			return config, fmt.Errorf("Limit must be set > 1 for bucket: %s", name)
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

package main

import (
	"gopkg.in/v1/yaml"
	"io/ioutil"
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

func NewConfiguration(path string) Configuration {

	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	config := Configuration{}
	yaml.Unmarshal(data, &config)

	return config
}

package config

import (
	"fmt"
	"github.com/Clever/sphinx/yaml"
	"io/ioutil"
)

// NewConfiguration takes in a path to a configuration yaml and returns a Configuration.
func NewConfiguration(path string) (yaml.Config, error) {

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return yaml.Config{},
			fmt.Errorf("failed to read %s. Aborting with error: %s", path, err.Error())
	}
	config, err := yaml.LoadAndValidateYaml(data)
	if err != nil {
		return yaml.Config{}, err
	}
	return config, nil
}

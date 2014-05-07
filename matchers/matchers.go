package matchers

import (
	"fmt"
	"github.com/Clever/sphinx/common"
	"gopkg.in/v1/yaml"
)

var (
	matcherFactories = [...]MatcherFactory{
		pathMatcherFactory{},
		headerMatcherFactory{},
	}
)

// MatcherFactoryFinder finds a MatcherFactory by name.
func MatcherFactoryFinder(name string) MatcherFactory {
	for _, factory := range matcherFactories {
		if factory.Type() == name {
			return factory
		}
	}
	return nil
}

type errorMatcherConfig struct {
	name    string
	message string
}

func (emc errorMatcherConfig) Error() string {
	return fmt.Sprintf("InvalidMatcherConfig: %s - %s",
		emc.name, emc.message)
}

// A Matcher determines if a common.Request matches its requirements.
type Matcher interface {
	Match(common.Request) bool
}

// A MatcherFactory creates a Matcher based on a config.
type MatcherFactory interface {
	Type() string
	Create(config interface{}) (Matcher, error)
}

// ReMarshal parses interface{} into concrete types
func ReMarshal(config interface{}, target interface{}) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, target)
}

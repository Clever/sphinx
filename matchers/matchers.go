package matchers

import (
	"github.com/Clever/sphinx/common"
	"gopkg.in/v1/yaml"
)

var matcherFactories = [...]MatcherFactory{
	PathMatcherFactory{},
	HeaderMatcherFactory{},
}

func MatcherFactoryFinder(name string) MatcherFactory {
	for _, factory := range matcherFactories {
		if factory.Type() == name {
			return factory
		}
	}
	return nil
}

type Matcher interface {
	Match(common.Request) bool
}

type MatcherFactory interface {
	Type() string
	Create(config interface{}) (Matcher, error)
}

func ReMarshal(config interface{}, target interface{}) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, target)
	if err != nil {
		return err
	}
	return nil
}

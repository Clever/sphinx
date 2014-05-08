package matchers

import (
	"fmt"
	"github.com/Clever/sphinx/common"
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

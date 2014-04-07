package matchers

import (
	"github.com/Clever/sphinx/common"
	"regexp"
)

type PathMatcherConfig struct {
	MatchAny []string `yaml:"match_any"`
}

type PathMatcher struct {
	Paths []*regexp.Regexp
}

func (pm PathMatcher) Match(request common.Request) bool {

	return false
}

type PathMatcherFactory struct{}

func (pmf PathMatcherFactory) Type() string {
	return "paths"
}

func (pmf PathMatcherFactory) Create(config interface{}) (Matcher, error) {
	matcherConfig := PathMatcherConfig{}
	var matcher PathMatcher

	err := ReMarshal(config, matcherConfig)
	if err != nil {
		return matcher, err
	}

	for _, p := range matcherConfig.MatchAny {
		compiled, err := regexp.Compile(p)
		if err != nil {
			return matcher, err
		}
		matcher.Paths = append(matcher.Paths, compiled)
	}
	return matcher, nil
}

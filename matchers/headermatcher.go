package matchers

import (
	"github.com/Clever/sphinx/common"
	"regexp"
)

type HeaderMatcherConfig struct {
	Name  string
	Match []string
}

type HeaderMatch struct {
	Name  string
	Match regexp.Regexp
}

type HeaderMatcher struct {
	Headers []HeaderMatch
}

func (hm HeaderMatcher) Match(request common.Request) bool {
	return false
}

type HeaderMatcherFactory struct{}

func (hmf HeaderMatcherFactory) Type() string {
	return "headers"
}

func (hmf HeaderMatcherFactory) Create(config interface{}) (Matcher, error) {
	var matcherConfig []HeaderMatcherConfig
	var matcher HeaderMatcher

	err := ReMarshal(config, matcherConfig)
	if err != nil {
		return matcher, err
	}

	return matcher, nil
}

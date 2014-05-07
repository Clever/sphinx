package matchers

import (
	"github.com/Clever/sphinx/common"
	"regexp"
)

type pathMatcherConfig struct {
	MatchAny []string `yaml:"match_any"`
}

type pathMatcher struct {
	Paths []*regexp.Regexp
}

func (pm pathMatcher) Match(request common.Request) bool {

	if _, ok := request["path"]; !ok {
		return false
	}

	// consider it a match if any of the headers match
	for _, matcher := range pm.Paths {
		if matcher.MatchString(request["path"].(string)) {
			return true
		}
	}

	return false
}

type pathMatcherFactory struct{}

func (pmf pathMatcherFactory) Type() string {
	return "paths"
}

func (pmf pathMatcherFactory) Create(config interface{}) (Matcher, error) {
	matcherConfig := pathMatcherConfig{}
	matcher := pathMatcher{}
	if err := reMarshal(config, &matcherConfig); err != nil {
		return matcher, err
	}

	if len(matcherConfig.MatchAny) == 0 {
		return matcher, errorMatcherConfig{
			name:    pmf.Type(),
			message: "missing key match_any or no paths",
		}
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

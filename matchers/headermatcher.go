package matchers

import (
	_ "fmt"
	"github.com/Clever/sphinx/common"
	"net/http"
	"regexp"
)

type HeaderMatcherConfig struct {
	MatchAny []HeaderMatchConfig `yaml:"match_any"`
}

type HeaderMatchConfig struct {
	Name  string
	Match string
}

type HeaderMatch struct {
	Name  string
	Match *regexp.Regexp
}

func NewHeaderMatch(name string, value string) (HeaderMatch, error) {

	var err error
	matcher := HeaderMatch{}
	matcher.Name = name
	if value == "" {
		matcher.Match = nil
	} else {
		matcher.Match, err = regexp.Compile(value)
	}

	return matcher, err
}

type HeaderMatcher struct {
	Headers []HeaderMatch
}

func (hm HeaderMatcher) Match(request common.Request) bool {

	// should have an header with hm.Name
	if _, ok := request["headers"]; !ok {
		return false
	}
	headers := request["headers"].(http.Header)
	for _, matcher := range hm.Headers {
		if header, ok := headers[http.CanonicalHeaderKey(matcher.Name)]; ok {
			// call it a match if there is no regexp
			if matcher.Match == nil {
				return true
			}
			// consider it a match when any of the header values pass the regexp
			for _, headerval := range header {
				if matcher.Match.MatchString(headerval) {
					return true
				}
			}
		}
	}

	return false
}

type HeaderMatcherFactory struct{}

func (hmf HeaderMatcherFactory) Type() string {
	return "headers"
}

func (hmf HeaderMatcherFactory) Create(config interface{}) (Matcher, error) {
	var matcherConfig HeaderMatcherConfig
	var matcher HeaderMatcher

	err := ReMarshal(config, &matcherConfig)
	if err != nil {
		return matcher, err
	}

	var headers []HeaderMatch
	for _, headerdetails := range matcherConfig.MatchAny {
		headermatch, err := NewHeaderMatch(headerdetails.Name,
			headerdetails.Match)
		if err != nil {
			return matcher, err
		}
		headers = append(headers, headermatch)
	}
	matcher.Headers = headers
	return matcher, nil
}

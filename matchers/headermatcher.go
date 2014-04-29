package matchers

import (
	_ "fmt"
	"github.com/Clever/sphinx/common"
	"net/http"
	"regexp"
)

// type to deserialize configuration
type matcherConfig struct {
	MatchAny []matcherItem `yaml:"match_any"`
}

type matcherItem struct {
	Name  string
	Match string
}

type HeaderMatcher struct {
	headers []headerMatch
}

type headerMatch struct {
	Name  string
	Match *regexp.Regexp
}

func NewHeaderMatch(name string, value string) (headerMatch, error) {

	matcher := headerMatch{Name: name}
	if value == "" {
		return matcher, nil
	}

	match, err := regexp.Compile(value)
	matcher.Match = match
	return matcher, err
}

func (hm HeaderMatcher) Match(request common.Request) bool {

	// should have an header with hm.Name
	if _, ok := request["headers"]; !ok {
		return false
	}
	headers := request["headers"].(http.Header)
	for _, matcher := range hm.headers {
		header, ok := headers[http.CanonicalHeaderKey(matcher.Name)]
		// ignore if header is not found
		if !ok {
			continue
		}
		// call it a match when header is found and there is no regexp
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
	return false
}

type HeaderMatcherFactory struct{}

func (hmf HeaderMatcherFactory) Type() string {
	return "headers"
}

func (hmf HeaderMatcherFactory) Create(config interface{}) (Matcher, error) {
	var headermatcherconfig matcherConfig

	err := ReMarshal(config, &headermatcherconfig)
	if err != nil {
		return HeaderMatcher{}, err
	}

	if len(headermatcherconfig.MatchAny) == 0 {
		return HeaderMatcher{}, ErrorMatcherConfig{
			name:    hmf.Type(),
			message: "missing key match_any",
		}
	}

	var headers []headerMatch
	for _, headerdetails := range headermatcherconfig.MatchAny {
		if headerdetails.Name == "" {
			return HeaderMatcher{}, ErrorMatcherConfig{
				name:    hmf.Type(),
				message: "name required for headers",
			}
		}

		headermatch, err := NewHeaderMatch(headerdetails.Name,
			headerdetails.Match)
		if err != nil {
			return HeaderMatcher{}, err
		}
		headers = append(headers, headermatch)
	}
	return HeaderMatcher{headers: headers}, nil
}

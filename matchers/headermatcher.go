package matchers

import (
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

type headerMatcher struct {
	headers []headerMatch
}

type headerMatch struct {
	Name  string
	Match *regexp.Regexp
}

func newHeaderMatch(name string, value string) (headerMatch, error) {

	matcher := headerMatch{Name: name}
	if value == "" {
		return matcher, nil
	}

	match, err := regexp.Compile(value)
	matcher.Match = match
	return matcher, err
}

func (hm headerMatcher) Match(request common.Request) bool {

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

type headerMatcherFactory struct{}

func (hmf headerMatcherFactory) Type() string {
	return "headers"
}

func (hmf headerMatcherFactory) Create(config interface{}) (Matcher, error) {
	headermatcherconfig := matcherConfig{}
	if err := reMarshal(config, &headermatcherconfig); err != nil {
		return headerMatcher{}, err
	}

	if len(headermatcherconfig.MatchAny) == 0 {
		return headerMatcher{}, errorMatcherConfig{
			name:    hmf.Type(),
			message: "missing key match_any",
		}
	}

	headers := []headerMatch{}
	for _, headerdetails := range headermatcherconfig.MatchAny {
		if headerdetails.Name == "" {
			return headerMatcher{}, errorMatcherConfig{
				name:    hmf.Type(),
				message: "name required for headers",
			}
		}

		headermatch, err := newHeaderMatch(headerdetails.Name, headerdetails.Match)
		if err != nil {
			return headerMatcher{}, err
		}
		headers = append(headers, headermatch)
	}
	return headerMatcher{headers: headers}, nil
}

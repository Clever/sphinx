package matchers

import (
	"github.com/Clever/sphinx/common"
	"gopkg.in/v1/yaml"
	"log"
	"testing"
)

type TestPathMatcherConfig struct {
	Paths interface{}
}

func getPathMatcher(config []byte) Matcher {
	var pathConfig TestPathMatcherConfig
	yaml.Unmarshal(config, &pathConfig)

	factory := PathMatcherFactory{}
	pathmatcher, err := factory.Create(pathConfig.Paths)
	if err != nil {
		log.Panicf("Failed to create PathMatcher", err)
	}

	return pathmatcher
}

func getRequestForPath(path string) common.Request {
	return map[string]interface{}{
		"path":       path,
		"headers":    nil,
		"remoteaddr": nil,
	}
}

func TestPathMatcherFactory(t *testing.T) {
	config := []byte(`
paths:
  match_any:
  - "/v1.1/push/events/.*"
  - "/v2.1/.*/events$"
`)

	pathmatcher := getPathMatcher(config)

	if len(pathmatcher.(PathMatcher).Paths) != 2 {
		log.Panicf("Expected two regexps in PathMatcher. Found: %d",
			len(pathmatcher.(PathMatcher).Paths))
	}
}

func TestPathMatcher(t *testing.T) {
	config := []byte(`
paths:
  match_any:
  - "/v1.1/push/events/.*"
  - "/v1.1/.*/events$"
`)

	pathmatcher := getPathMatcher(config)
	request := getRequestForPath("/v1.1/push/events/students/12234234")
	if !pathmatcher.Match(request) {
		log.Panicf("Expected event request to match config")
	}

	request = getRequestForPath("/v1.1/students/12234234/events")
	if !pathmatcher.Match(request) {
		log.Panicf("Expected event request to match config")
	}

	request = getRequestForPath("/v1.1/students/12234234")
	if pathmatcher.Match(request) {
		log.Panicf("Do not expect resource request to match config")
	}
}

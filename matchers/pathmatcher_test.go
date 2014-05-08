package matchers

import (
	"fmt"
	"github.com/Clever/sphinx/common"
	"gopkg.in/v1/yaml"
	"log"
	"strconv"
	"testing"
)

type TestPathMatcherConfig struct {
	Paths interface{}
}

func getPathMatcher(config []byte) pathMatcher {
	pathConfig := TestPathMatcherConfig{}
	yaml.Unmarshal(config, &pathConfig)

	factory := pathMatcherFactory{}
	pathmatcher, err := factory.Create(pathConfig.Paths)
	if err != nil {
		log.Panicf("Failed to create PathMatcher", err)
	}

	return pathmatcher.(pathMatcher)
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

	if len(pathmatcher.Paths) != 2 {
		log.Panicf("Expected two regexps in PathMatcher. Found: %d",
			len(pathmatcher.Paths))
	}
}

func TestPathMatcherFactoryBadConfig(t *testing.T) {
	config := []byte(`
paths:
  - "/v1.1/push/events/.*"
  - "/v2.1/.*/events$"
`)
	pathConfig := TestPathMatcherConfig{}
	yaml.Unmarshal(config, &pathConfig)

	factory := pathMatcherFactory{}
	_, err := factory.Create(pathConfig.Paths)
	if err == nil {
		t.Error("Expected error when headers have no name")
	}

	config = []byte(`
paths:
  match_any: "hello"
`)
	yaml.Unmarshal(config, &pathConfig)
	_, err = factory.Create(pathConfig.Paths)
	if err == nil {
		t.Error("Expected error when headers have no name")
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

// Benchmarks PathMatcher.Match with a config with numPaths paths and
// requests with numPaths paths, where none of the paths match.
var benchPath = func(b *testing.B, numPaths int) {
	config := "paths:\n  match_any:\n"
	for i := 0; i < numPaths; i++ {
		str := strconv.Itoa(i)
		config += fmt.Sprintf("    - \"/v1.1/path/%s\"\n", str)
	}
	pathMatcher := getPathMatcher([]byte(config))
	request := getRequestForPath("/v1.1/path/" + strconv.Itoa(numPaths))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pathMatcher.Match(request)
	}
}

func Benchmark1Path(b *testing.B) {
	benchPath(b, 1)
}

func Benchmark100Paths(b *testing.B) {
	benchPath(b, 100)
}

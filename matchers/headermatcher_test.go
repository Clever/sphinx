package matchers

import (
	"github.com/Clever/sphinx/common"
	"gopkg.in/v1/yaml"
	"testing"
)

type TestHeaderConfig struct {
	Headers interface{}
}

func getHeaderMatcher(config []byte) (Matcher, error) {

	var headerConfig TestHeaderConfig
	yaml.Unmarshal(config, &headerConfig)

	factory := HeaderMatcherFactory{}
	return factory.Create(headerConfig.Headers)
}

func getRequest(headers map[string][]string) common.Request {
	httprequest := common.ConstructMockRequestWithHeaders(headers)
	return common.HttpToSphinxRequest(httprequest)
}

func TestHeaderMatcherFactory(t *testing.T) {
	config := []byte(`
headers:
  match_any:
    - name: "Authorization"
      match: "Bearer.*"
    - name: "X-Forwarded-For"
`)
	headermatcher, err := getHeaderMatcher(config)
	if err != nil {
		t.Fatalf("Failed to create HeaderMatcher", err)
	}

	if len(headermatcher.(HeaderMatcher).headers) != 2 {
		t.Fatalf("Expected two Headers in HeaderMatcher found: %d",
			len(headermatcher.(HeaderMatcher).headers))
	}
	for _, header := range headermatcher.(HeaderMatcher).headers {
		if header.Name == "X-Forwarded-For" {
			if header.Match != nil {
				t.Fatalf("Expected X-Forwarded-For match to be nil. Found:%s",
					header.Match.String())
			}
		} else {
			if header.Match == nil {
				t.Fatalf("Expected Authorization header to have a match")
			}
		}
	}
}

func TestHeaderMatcherFactoryBadData(t *testing.T) {
	config := []byte(`
headers:
  match_any:
    - "Authorization": "Bearer.*"
    - name: "X-Forwarded-For"
`)
	var headerConfig TestHeaderConfig
	yaml.Unmarshal(config, &headerConfig)

	factory := HeaderMatcherFactory{}
	_, err := factory.Create(headerConfig.Headers)
	if err == nil {
		t.Error("Expected error when headers have no name")
	}

	config = []byte(`
headers:
  - "Authorization": "Bearer.*"
  - name: "X-Forwarded-For"
`)
	yaml.Unmarshal(config, &headerConfig)
	_, err = factory.Create(headerConfig.Headers)
	if err == nil {
		t.Error("expected error when match_any is missing")
	}
}

func TestRegexMatches(t *testing.T) {
	config := []byte(`
headers:
  match_any:
    - name: "Authorization"
      match: "Bearer.*"
    - name: "X-Forwarded-For"
      match: "192.0.0.1"
`)
	headermatcher, err := getHeaderMatcher(config)
	if err != nil {
		t.Fatalf("Failed to create HeaderMatcher", err)
	}
	request := getRequest(map[string][]string{
		"Authorization": []string{"Bearer 12345"},
	})

	if !headermatcher.Match(request) {
		t.Fatalf("Should have matched Header Authorization")
	}

	request = getRequest(map[string][]string{
		"Authorization": []string{"Basic 12345"},
	})
	if headermatcher.Match(request) {
		t.Fatalf("Should NOT have matched Header Authorization Basic")
	}

	request = getRequest(map[string][]string{
		"X-Forwarded-For": []string{"192.0.0.1", "127.0.0.1"},
		"Authorization":   []string{"Basic 12345"},
	})
	if !headermatcher.Match(request) {
		t.Fatalf("Should have matched X-Forwarded-For")
	}
	request = getRequest(map[string][]string{
		"Authorization": []string{"Basic 12345"},
	})
	if headermatcher.Match(request) {
		t.Fatalf("Should NOT have matched Header Authorization Basic")
	}
}

func TestHeaderPresence(t *testing.T) {
	config := []byte(`
headers:
  match_any:
    - name: "Authorization"
`)
	headermatcher, err := getHeaderMatcher(config)
	if err != nil {
		t.Fatalf("Failed to create HeaderMatcher", err)
	}

	request := getRequest(map[string][]string{
		"Authorization": []string{"Bearer 12345"},
	})
	if !headermatcher.Match(request) {
		t.Fatalf("Should have matched Header Authorization")
	}

	request = getRequest(map[string][]string{
		"X-Forwarded-For": []string{"192.0.0.1"},
	})
	if headermatcher.Match(request) {
		t.Fatalf("Should NOT have matched Header X-Forwarded-For")
	}
}

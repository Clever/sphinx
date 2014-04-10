package matchers

import (
	//"github.com/Clever/sphinx"
	"github.com/Clever/sphinx/common"
	"gopkg.in/v1/yaml"
	"log"
	"testing"
)

type TestHeaderConfig struct {
	Headers interface{}
}

func TestHeaderMatcherFactory(t *testing.T) {

	config := []byte(`
headers:
  match_any:
    - name: "Authorization"
      match: "Bearer.*"
    - name: "X-Forwarded-For"
`)
	var headerConfig TestHeaderConfig
	yaml.Unmarshal(config, &headerConfig)

	factory := HeaderMatcherFactory{}
	headermatcher, err := factory.Create(headerConfig.Headers)

	if err != nil {
		log.Panicf("Failed to create HeaderMatcher", err)
	}
	if len(headermatcher.(HeaderMatcher).Headers) != 2 {
		log.Panicf("Expected two Headers in HeaderMatcher found: %d",
			len(headermatcher.(HeaderMatcher).Headers))
	}
	for _, header := range headermatcher.(HeaderMatcher).Headers {
		if header.Name == "X-Forwarded-For" {
			if header.Match != nil {
				log.Panicf("Expected X-Forwarded-For match to be nil. Found:%s",
					header.Match.String())
			}
		} else {
			if header.Match == nil {
				log.Panicf("Expected Authorization header to have a match")
			}
		}
	}

	httprequest := common.ConstructMockRequestWithHeaders(map[string][]string{
		"Authorization": []string{"Bearer 12345"},
	})
	request := common.HttpToSphinxRequest(httprequest)

	if !headermatcher.Match(request) {
		log.Panicf("Should have matched Header Authorization")
	}
}

package limitkeys

import (
	"github.com/Clever/sphinx/common"
	"log"
	"testing"
)

func getRequest(headers map[string][]string) common.Request {
	httprequest := common.ConstructMockRequestWithHeaders(headers)
	return common.HttpToSphinxRequest(httprequest)
}

func TestHeaderLimitKey(t *testing.T) {
	limitkey := HeaderLimitKey{
		name: "Authorization",
	}

	request := getRequest(map[string][]string{
		"Authorization": []string{"Bearer 12345"},
	})
	if key, err := limitkey.Key(request); err != nil || key != "Authorization:Bearer 12345" {
		log.Panicf("HeaderKey did not match")
	}

	// returns custom error when no key is found
	request = getRequest(map[string][]string{
		"X-Forwarded-For": []string{"127.0.0.1"},
	})
	key, err := limitkey.Key(request)
	if key != "" {
		log.Panicf("Expecting empty key when header is not found, but got: %s", key)
	}
	if emptyKeyError, ok := err.(EmptyKeyError); !ok ||
		emptyKeyError.Error() != "LimitKeyType: header. Header Authorization not found in request" {

		log.Panicf("Failed to return correct error when required Header is not found")
	}

	// works with arrays in headers
	// currently creates a new key for change in any one of the array elements
	// i.e. Keys are created by concatenating the array elements
	limitkey = HeaderLimitKey{
		name: "X-Forwarded-For",
	}

	request = getRequest(map[string][]string{
		"X-Forwarded-For": []string{"127.0.0.1", "172.0.0.1"},
	})
	if key, err := limitkey.Key(request); err != nil || key !=
		"X-Forwarded-For:127.0.0.1;172.0.0.1" {
		log.Panicf("Header key for X-Forwarded-For did not match")
	}
}

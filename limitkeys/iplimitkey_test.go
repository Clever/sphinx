package limitkeys

import (
	"log"
	"testing"
)

func TestIPLimitKey(t *testing.T) {
	limitkey := IPLimitKey{}

	request := map[string]interface{}{
		"remoteaddr": "127.0.0.1",
	}
	if key, err := limitkey.Key(request); err != nil || key != "ip:127.0.0.1" {
		log.Panicf("IPLimitKey did not match")
	}

	// returns correct error when remoteaddr is not set
	request = map[string]interface{}{
		"headers": "boo",
	}
	key, err := limitkey.Key(request)
	if key != "" {
		log.Panicf("Expecting empty key when remoteaddr is not found, but got: %s", key)
	}
	if emptyKeyError, ok := err.(EmptyKeyError); !ok ||
		emptyKeyError.Error() != "LimitKeyType: ip. No remoteaddr key in request" {

		log.Panicf("Failed to return correct error when ip is not found")
	}
}

package limitkeys

import (
	"testing"
)

func TestIPLimitKey(t *testing.T) {
	limitkey := ipLimitKey{}

	request := map[string]interface{}{
		"remoteaddr": "127.0.0.1",
	}
	if key, err := limitkey.Key(request); err != nil || key != "ip:127.0.0.1" {
		t.Error("IPLimitKey did not match")
	}

}

// returns correct error when remoteaddr is not set
func TestIPLimitKeyNoRemoteAddr(t *testing.T) {
	limitkey := ipLimitKey{}

	request := map[string]interface{}{
		"headers": "boo",
	}
	key, err := limitkey.Key(request)
	if key != "" {
		t.Errorf("Expecting empty key when remoteaddr is not found, but got: %s", key)
	}
	if emptyKeyError, ok := err.(EmptyKeyError); !ok ||
		emptyKeyError.Error() != "LimitKeyType: ip. No remoteaddr key in request" {
		t.Error("Failed to return correct error when ip is not found")
	}
}

package limitkeys

import (
	"testing"
)

// returns the same key all the time
func TestGlobalLimitKey(t *testing.T) {
	limitkey := globalLimitKey{}

	request := map[string]interface{}{}
	key, err := limitkey.Key(request)
	if err != nil || key != globalLimitKeyValue {
		t.Errorf("Key %s did not match default global key", key)
	}

	request = map[string]interface{}{
		"headers":   "boo",
		"something": "does-not-matter",
	}
	key, err = limitkey.Key(request)
	if err != nil || key != globalLimitKeyValue {
		t.Errorf("Key %s did not match default global key", key)
	}
}

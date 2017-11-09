package common

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"sort"

	"gopkg.in/yaml.v2"
)

// Request contains any info necessary to ratelimit a request
type Request map[string]interface{}

// InSlice tests whether or not a string exists in a slice of strings
func InSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// ReMarshal parses interface{} into concrete types
func ReMarshal(config interface{}, target interface{}) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, target)
}

// HTTPToSphinxRequest converts an http.Request to a Request
func HTTPToSphinxRequest(r *http.Request) Request {
	return map[string]interface{}{
		"path":       r.URL.Path,
		"headers":    r.Header,
		"remoteaddr": r.RemoteAddr,
		"method":     r.Method,
	}
}

// Hash hashes a string based on the given salt
func Hash(str, salt string) string {
	if salt == "" {
		return str
	}
	hash := hmac.New(sha256.New, []byte(salt))
	hash.Write([]byte(str))
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

// SortedKeys returns a sorted slice of map keys
func SortedKeys(obj map[string]interface{}) []string {
	// use make so as to prevent re-allocation
	keys := make([]string, len(obj))
	i := 0
	for k := range obj {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

package common

import (
	"gopkg.in/v1/yaml"
	"net/http"
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
	}
}

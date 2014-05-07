package common

import (
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

// HTTPToSphinxRequest converts an http.Request to a Request
func HTTPToSphinxRequest(r *http.Request) Request {
	return map[string]interface{}{
		"path":       r.URL.Path,
		"headers":    r.Header,
		"remoteaddr": r.RemoteAddr,
	}
}

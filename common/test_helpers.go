package common

import (
	"net/http"
	"net/url"
)

// HTTPToSphinxRequest converts an http.Request to a Request
func HTTPToSphinxRequest(r *http.Request) Request {
	return map[string]interface{}{
		"path":       r.URL.Path,
		"headers":    r.Header,
		"remoteaddr": r.RemoteAddr,
	}
}

// ConstructMockRequestWithHeaders constructs an http.Request with the given headers
func ConstructMockRequestWithHeaders(headers map[string][]string) *http.Request {
	testURL, err := url.Parse("https://google.com/trolling/path")
	if err != nil {
		panic(err)
	}

	return &http.Request{
		Header: headers,
		URL:    testURL,
	}
}

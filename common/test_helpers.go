package common

import (
	"net/http"
	"net/url"
)

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

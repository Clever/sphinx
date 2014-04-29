package common

import (
	"net/http"
	"net/url"
)

func HttpToSphinxRequest(r *http.Request) Request {
	return map[string]interface{}{
		"path":       r.URL.Path,
		"headers":    r.Header,
		"remoteaddr": r.RemoteAddr,
	}
}

func ConstructMockRequestWithHeaders(headers map[string][]string) *http.Request {
	testUrl, err := url.Parse("https://google.com/trolling/path")
	if err != nil {
		panic(err)
	}

	return &http.Request{
		Header: headers,
		URL:    testUrl,
	}
}

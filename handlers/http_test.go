package http

import (
	"net/http"
	"net/url"
	"testing"
)

func constructMockRequestWithHeaders(headers map[string][]string) *http.Request {
	testUrl, err := url.Parse("https://google.com/trolling/path")
	if err != nil {
		panic(err)
	}

	return &http.Request{
		Header: headers,
		URL:    testUrl,
	}
}

func compareStrings(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Fatalf("expected %s, received %s", expected, actual)
	}
}

func compareHeader(t *testing.T, headers http.Header, key string, values []string) {
	if len(headers[key]) != len(values) {
		t.Fatalf("expected %d '%s' headers, received %d", len(values), len(headers[key]))
	}
	for i, expected := range values {
		compareStrings(t, expected, headers[key][i])
	}
}

func TestParsesHeaders(t *testing.T) {
	request := parseRequest(constructMockRequestWithHeaders(map[string][]string{
		"Authorization":   []string{"Bearer 12345"},
		"X-Forwarded-For": []string{"IP1", "IP2"},
	}))
	if len(request["headers"].(http.Header)) != 2 {
		t.Fatalf("Expected 2 header, recevied %d", len(request["headers"].(http.Header)))
	}

	compareHeader(t, request["headers"].(http.Header), "Authorization", []string{"Bearer 12345"})
	compareHeader(t, request["headers"].(http.Header), "X-Forwarded-For", []string{"IP1", "IP2"})
	compareStrings(t, request["path"].(string), "/trolling/path")
}

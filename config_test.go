package main

import (
	"testing"
)

func TestNewConfiguration(t *testing.T) {
	config := NewConfiguration("./example.yaml")

	if config.Forward.Scheme != "http" {
		t.Error("expected http for Forward.Scheme")
	}

	if len(config.Buckets) != 4 {
		t.Error("expected 4 bucket definitions")
	}

	for _, bucket := range config.Buckets {
		if bucket.Interval < 1 {
			t.Error("bucket interval should be greator than 1")
		}
		if bucket.Limit < 1 {
			t.Error("bucket limit should be greator than 1")
		}
		if len(bucket.Keys) < 1 {
			t.Error("bucket was expected to have 1 key")
		}

		if len(bucket.Matches.Headers) < 1 && len(bucket.Matches.Paths) < 1 {
			t.Error("One of paths or headers was expected to be set for matches")
		}
	}
}

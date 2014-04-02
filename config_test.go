package sphinx

import (
	"testing"
)

func TestNewConfiguration(t *testing.T) {

	// test loading example config
	config := NewConfiguration("./example.yaml")

	if config.Forward.Scheme != "http" {
		t.Error("expected http for Forward.Scheme")
	}

	if len(config.Limits) != 4 {
		t.Error("expected 4 bucket definitions")
	}

	for _, limit := range config.Limits {
		if limit.Interval < 1 {
			t.Error("limit interval should be greator than 1")
		}
		if limit.Max < 1 {
			t.Error("limit max should be greator than 1")
		}
		if len(limit.Keys) < 1 {
			t.Error("limit was expected to have 1 key")
		}

		if len(limit.Matches["headers"]) < 1 && len(limit.Matches["paths"]) < 1 {
			t.Error("One of paths or headers was expected to be set for matches")
		}
	}

	// test incorrect config
	invalid_config := []byte(`
forward:
  host: proxy.example.com
`)
	config, err := loadAndValidateConfig(invalid_config)
	if err == nil {
		t.Error("invalid config did not return error")
	}

	invalid_config = []byte(`
forward:
  scheme: http
  host: proxy.example.com

buckets:
  bearer/events:
    keys:
      - 'header:authentication'
`)
	config, err = loadAndValidateConfig(invalid_config)
	if err == nil {
		t.Error("invalid config did not return error")
	}
}

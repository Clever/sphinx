package sphinx

import (
	"testing"
)

// test example config file is loaded correctly
func TestConfigurationFileLoading(t *testing.T) {

	config, err := NewConfiguration("./example.yaml")
	if err != nil {
		t.Error("could not load example configuration")
	}

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
		if limit.Keys != nil {
			t.Error("limit was expected to have atleast 1 key")
		}

		if limit.Matches["headers"] == nil && limit.Matches["paths"] == nil {
			t.Error("One of paths or headers was expected to be set for matches")
		}
	}
}

// Incorrect configuration file should return errors
func TestConfigurationFileFailures(t *testing.T) {

	invalid_config := []byte(`
forward:
  host: proxy.example.com
`)
	config, err = loadAndValidateConfig(invalid_config)
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

package config

import (
	"strings"
	"testing"
)

// test example config file is loaded correctly
func TestConfigurationFileLoading(t *testing.T) {

	config, err := NewConfiguration("../example.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Error("could not load example configuration")
	}

	if config.Proxy.Handler != "http" {
		t.Error("expected http for Proxy.Handler")
	}

	if len(config.Limits) != 4 {
		t.Error("expected 4 bucket definitions")
	}

	for name, limit := range config.Limits {
		if limit.Interval < 1 {
			t.Errorf("limit interval should be greator than 1 for limit: %s", name)
		}
		if limit.Max < 1 {
			t.Errorf("limit max should be greator than 1 for limit: %s", name)
		}
	}
}

func TestInvalidConfigurationPath(t *testing.T) {
	if _, err := NewConfiguration("./does-not-exist.yaml"); err == nil {
		t.Fatalf("Expected error for invalid config path")
	} else if !strings.Contains(err.Error(), "no such file or directory") {
		t.Fatalf("Expected no file error got %s", err.Error())
	}
}

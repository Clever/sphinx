package sphinx

import (
	"bytes"
	"strings"
	"testing"
)

// test example config file is loaded correctly
func TestConfigurationFileLoading(t *testing.T) {

	config, err := NewConfiguration("./example.yaml")
	if err != nil {
		t.Error("could not load example configuration")
	}

	if config.Proxy.Handler != "http" {
		t.Error("expected http for Proxy.Handler")
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

func TestInvalidConfigurationPath(t *testing.T) {
	_, err := NewConfiguration("./does-not-exist.yaml")
	if err == nil {
		t.Error("Expected error for invalid config path")
	}
	if !strings.Contains(err.Error(), "no such file or directory") {
		t.Errorf("Expected no file error got %s", err.Error())
	}
}

func TestInvalidYaml(t *testing.T) {
	invalidYaml := []byte(`
forward
  host$$: proxy.example.com
`)
	_, err := loadAndValidateConfig(invalidYaml)
	if !strings.Contains(err.Error(), "YAML error:") {
		t.Errorf("expected yaml error, got %s", err.Error())
	}
}

// Incorrect configuration file should return errors
func TestInvalidProxyConfig(t *testing.T) {

	invalidConfig := []byte(`
proxy:
  host: http://proxy.example.com
`)
	_, err := loadAndValidateConfig(invalidConfig)
	if err == nil || !strings.Contains(err.Error(), "handler") {
		t.Errorf("Expected proxy handler error. Got: %s", err.Error())
	}

	invalidConfig = []byte(`
proxy:
  handler: http
  host: proxy.example.com
`)
	_, err = loadAndValidateConfig(invalidConfig)
	if err == nil || !strings.Contains(err.Error(), "host:port") {
		t.Errorf("Expected proxy host error. Got: %s", err.Error())
	}

	invalidConfig = []byte(`
proxy:
  handler: http
  host: proxy.example.com
  listen: :8000
`)
	_, err = loadAndValidateConfig(invalidConfig)
	if err == nil || !strings.Contains(err.Error(), "proxy") {
		t.Errorf("Expected proxy host error. Got: %s", err.Error())
	}
}

func TestInvalidLimitConfig(t *testing.T) {

	baseBuf := bytes.NewBufferString(`
proxy:
  handler: http
  host: http://proxy.example.com
  listen: "0.0.0.0:8080"
storage:
  type: memory
`)

	configBuf := baseBuf
	configBuf.WriteString(`
limits:
  bearer/events:
    keys:
      - 'header:authentication'
`)
	_, err := loadAndValidateConfig(configBuf.Bytes())
	if err == nil || !strings.Contains(err.Error(), "interval") {
		t.Errorf("Expected Limit Interval error. Got: %s", err.Error())
	}

	configBuf = baseBuf
	configBuf.WriteString(`
limits:
  bearer/events:
    interval: 10
    keys:
      - 'header:authentication'
`)
	_, err = loadAndValidateConfig(configBuf.Bytes())
	if err == nil || !strings.Contains(err.Error(), "max") {
		t.Errorf("Expected Limit Interval error. Got: %s", err.Error())
	}
}

func TestInvalidStorageConfig(t *testing.T) {
	baseBuf := bytes.NewBufferString(`
proxy:
  handler: http
  host: http://proxy.example.com
  listen: localhost:8080
limits:
  test:
    interval: 15  # in seconds
    max: 200
`)

	_, err := loadAndValidateConfig(baseBuf.Bytes())
	if err == nil || !strings.Contains(err.Error(), "storage type must be set") {
		t.Errorf("Expected Storage error. Got: %s", err.Error())
	}

	configBuf := baseBuf
	configBuf.WriteString(`
storage:
  type: redis
`)
	_, err = loadAndValidateConfig(configBuf.Bytes())
	if err == nil || !strings.Contains(err.Error(), "host") {
		t.Errorf("Expected redis Storage host error. Got: %s", err.Error())
	}

	configBuf = baseBuf
	configBuf.WriteString(`
storage:
  type: redis
  host: localhost
`)
	_, err = loadAndValidateConfig(configBuf.Bytes())
	if err == nil || !strings.Contains(err.Error(), "port") {
		t.Errorf("Expected redis Storage host error. Got: %s", err.Error())
	}
}

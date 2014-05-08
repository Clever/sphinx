package config

import (
	"bytes"
	"strings"
	"testing"
)

// test that matcher errors are bubbled up
func TestBadConfiguration(t *testing.T) {

	configBuf := bytes.NewBufferString(`
proxy:
  handler: http
  host: http://proxy.example.com
  listen: :8080
storage:
  type: memory
limits:
  test:
    interval: 15  # in seconds
    max: 200
`)

	// header matchers are verified
	configBuf.WriteString(`
    keys:
      headers:
        - Authorization
    matches:
      headers:
        match_any:
          - "Authorization": "Bearer.*"
          - name: "X-Forwarded-For"
`)
	config, err := loadAndValidateYaml(configBuf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	if _, err := parseYaml(config); err == nil {
		t.Fatal("expected error")
	} else if !strings.Contains(err.Error(), "InvalidMatcherConfig: headers") {
		t.Errorf("Expected a InvalidMatcherConfig error, got different error: %s", err.Error())
	}

}

func TestInvalidYaml(t *testing.T) {
	invalidYaml := []byte(`
forward
  host$$: proxy.example.com
`)

	if _, err := loadAndValidateYaml(invalidYaml); !strings.Contains(err.Error(), "YAML error:") {
		t.Errorf("expected yaml error, got %s", err.Error())
	}
}

// Incorrect configuration file should return errors
func TestInvalidProxyConfig(t *testing.T) {

	invalidConfig := []byte(`
proxy:
  host: http://proxy.example.com
`)
	_, err := loadAndValidateYaml(invalidConfig)
	if err == nil || !strings.Contains(err.Error(), "handler") {
		t.Errorf("Expected proxy handler error. Got: %s", err.Error())
	}

	invalidConfig = []byte(`
proxy:
  handler: http
  host: proxy.example.com
`)
	_, err = loadAndValidateYaml(invalidConfig)
	if err == nil || !strings.Contains(err.Error(), "host:port") {
		t.Errorf("Expected proxy host error. Got: %s", err.Error())
	}

	invalidConfig = []byte(`
proxy:
  handler: http
  host: proxy.example.com
  listen: :8000
`)
	_, err = loadAndValidateYaml(invalidConfig)
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
      headers: 
        - 'Authentication'
`)
	_, err := loadAndValidateYaml(configBuf.Bytes())
	if err == nil || !strings.Contains(err.Error(), "interval") {
		t.Errorf("Expected Limit Interval error. Got: %s", err.Error())
	}

	configBuf = baseBuf
	configBuf.WriteString(`
limits:
  bearer/events:
    interval: 10
    keys:
      headers: 
        - 'Authentication'
`)
	_, err = loadAndValidateYaml(configBuf.Bytes())
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
    keys:
      headers:
        - Authorization
    interval: 15  # in seconds
    max: 200
`)

	_, err := loadAndValidateYaml(baseBuf.Bytes())
	if err == nil || !strings.Contains(err.Error(), "storage type must be set") {
		t.Errorf("Expected Storage error. Got: %s", err.Error())
	}

	configBuf := baseBuf
	configBuf.WriteString(`
storage:
  type: redis
`)
	_, err = loadAndValidateYaml(configBuf.Bytes())
	if err == nil || !strings.Contains(err.Error(), "host") {
		t.Errorf("Expected redis Storage host error. Got: %s", err.Error())
	}

	configBuf = baseBuf
	configBuf.WriteString(`
storage:
  type: redis
  host: localhost
`)
	_, err = loadAndValidateYaml(configBuf.Bytes())
	if err == nil || !strings.Contains(err.Error(), "port") {
		t.Errorf("Expected redis Storage host error. Got: %s", err.Error())
	}
}

package daemon

import (
	"fmt"
	"github.com/Clever/sphinx/config"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestConfigReload(t *testing.T) {
	conf, err := config.New("../example.yaml")
	if err != nil {
		t.Fatal("Error loading config: " + err.Error())
	}
	d := daemon{}
	d.LoadConfig(conf)
	if d.handler == nil {
		t.Fatal("Didn't assign handler")
	}
}

func TestFailedReload(t *testing.T) {
	conf, err := config.New("../example.yaml")
	if err != nil {
		t.Fatal("Error loading config: " + err.Error())
	}
	daemon, err := New(conf)
	if err != nil {
		t.Fatal("Error creating new daemon: " + err.Error())
	}
	conf2 := config.Config{}
	err = daemon.LoadConfig(conf2)
	if err == nil {
		t.Fatal("Should have errored on empty configuration")
	}

	conf.Proxy.Listen = ":1000"
	err = daemon.LoadConfig(conf)
	if err == nil {
		t.Fatalf("Should have errored on changed listen port")
	}
}

func setUpDaemonWithLocalServer(conf config.Config) error {
	// Set up a local server that 404s everywhere except route '/healthyroute'.
	mux := http.NewServeMux()
	mux.HandleFunc("/healthyroute", func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte("healthy"))
	})
	mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte("404"))
	})

	// Start local server on the port config points to.
	colonIdx := strings.LastIndex(conf.Proxy.Host, ":")
	localServerListen := conf.Proxy.Host[colonIdx:]
	go http.ListenAndServe(localServerListen, mux)

	// Set up and start the daemon.
	daemon, err := New(conf)
	if err != nil {
		return err
	}

	go daemon.Start()
	return nil
}

// testProxyRequest calls the proxy server at the given path and verifies that
// the request returns the given HTTP status and body content.
func testProxyRequest(t *testing.T, url string, expectedStatus int, expectedBody string) {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Proxy request failed: %s", err.Error())
	}

	if resp.StatusCode != expectedStatus {
		t.Fatalf("Response status with url %s does not match expected value.  Actual: %d.  Expected: %d.",
			url, resp.StatusCode, expectedStatus)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Couldn't read response body: %s", err.Error())
	}

	bodyStr := string(body)
	if bodyStr != expectedBody {
		t.Fatalf("Response body does not match expected value. Actual: \"%s\" Expected: \"%s\"",
			bodyStr, expectedBody)
	}
}

func TestHealthCheck(t *testing.T) {
	// Set up the daemon config to proxy to the local server.
	conf, err := config.New("../example.yaml")
	if err != nil {
		t.Fatalf("Couldn't load daemon config: %s", err.Error())
	}
	conf.Proxy.Host = "http://localhost:8000"
	conf.Proxy.Listen = ":6634"
	conf.HealthCheck.Port = "60002"
	conf.HealthCheck.Enabled = true

	err = setUpDaemonWithLocalServer(conf)
	if err != nil {
		t.Fatalf("Test daemon setup failed: %s", err.Error())
	}

	localProxyURL := "http://localhost" + conf.Proxy.Listen

	// Test a route that should be proxied to 404.
	testProxyRequest(t, localProxyURL+"/helloworld", http.StatusNotFound, "404")

	// Test a route that should be proxied to a valid response.
	testProxyRequest(t, localProxyURL+"/healthyroute", http.StatusOK, "healthy")

	healthCheckURL := fmt.Sprintf("http://localhost:%s/health/check", conf.HealthCheck.Port)

	// Test the health check.
	testProxyRequest(t, healthCheckURL, http.StatusOK, "")
}

func TestDaemonWithNoHealthCheck(t *testing.T) {
	// Set up the daemon config to proxy to the local server.
	conf, err := config.New("../example.yaml")
	if err != nil {
		t.Fatalf("Couldn't load daemon config: %s", err.Error())
	}
	conf.Proxy.Host = "http://localhost:8001"
	conf.Proxy.Listen = ":6635"
	conf.HealthCheck.Port = "60003"
	conf.HealthCheck.Enabled = false

	err = setUpDaemonWithLocalServer(conf)
	if err != nil {
		t.Fatalf("Test daemon setup failed: %s", err.Error())
	}

	// Because so many servers are starting, sleep for a second to make sure
	// they start.
	time.Sleep(time.Second)

	localProxyURL := "http://localhost" + conf.Proxy.Listen

	// Test a route that should be proxied to 404.
	testProxyRequest(t, localProxyURL+"/helloworld", http.StatusNotFound, "404")

	// Test a route that should be proxied to a valid response.
	testProxyRequest(t, localProxyURL+"/healthyroute", http.StatusOK, "healthy")
}

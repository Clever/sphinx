package daemon

import (
	"github.com/Clever/sphinx/config"
	"io/ioutil"
	"net/http"
	"testing"
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

var localServerPort = ":8081"
var localServerHost = "http://localhost" + localServerPort
var localProxyHost = "http://localhost:6634"

func setUpDaemonWithLocalServer() error {
	// Set up a local server that 404s everywhere except one route.
	mux := http.NewServeMux()
	mux.HandleFunc("/healthyroute", func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte("healthy"))
	})
	mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte("404"))
	})
	go http.ListenAndServe(localServerPort, mux)

	// Set up the daemon to proxy to the local server.
	conf, err := config.New("../example.yaml")
	if err != nil {
		return err
	}
	conf.Proxy.Host = localServerHost

	daemon, err := New(conf)
	if err != nil {
		return err
	}

	go daemon.Start()
	return nil
}

// testProxyRequest calls the proxy server at the given path and verifies that
// the request returns the given HTTP status and body content.
func testProxyRequest(t *testing.T, path string, expectedStatus int, expectedBody string) {
	resp, err := http.Get(localProxyHost + path)
	if err != nil {
		t.Fatalf("Proxy request failed: %s", err.Error())
	}

	if resp.StatusCode != expectedStatus {
		t.Fatalf("Response status with path %s does not match expected value.  Actual: %d.  Expected: %d.",
			path, resp.StatusCode, expectedStatus)
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
	err := setUpDaemonWithLocalServer()
	if err != nil {
		t.Fatalf("DAEMON_SETUP_FAILED: %s", err.Error())
	}

	// Test a route that should be proxied to 404.
	testProxyRequest(t, "/helloworld", http.StatusNotFound, "404")

	// Test a route that should be proxied to a valid response.
	testProxyRequest(t, "/healthyroute", http.StatusOK, "healthy")

	// Test the health check.
	testProxyRequest(t, "/sphinx/health/check", http.StatusOK, "")
}

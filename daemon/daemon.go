package daemon

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/Clever/sphinx/common"
	"github.com/Clever/sphinx/config"
	"github.com/Clever/sphinx/handlers"
	"github.com/Clever/sphinx/ratelimiter"
	"github.com/pborman/uuid"
	"gopkg.in/Clever/kayvee-go.v3/middleware"
)

// Daemon represents a daemon server
type Daemon interface {
	Start()
	LoadConfig(config config.Config) error
}

type daemon struct {
	handler     http.Handler
	proxy       config.Proxy
	healthCheck config.HealthCheck
}

// setUpHealthCheckService sets up a health check service at the given port
// that can be pinged at the given endpoint to determine if Sphinx is still
// running.
func setUpHealthCheckService(port string, endpoint string) {
	mux := http.NewServeMux()
	mux.HandleFunc(endpoint, func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		common.Log.Counter("Health-Check")
	})
	go http.ListenAndServe(":"+port, mux)
	log.Printf("Health-check listening on:%s%s", port, endpoint)
}

func (d *daemon) Start() {
	log.Printf("Limiter listening on %s", d.proxy.Listen)
	// Only set up the health check service if it is enabled.
	if d.healthCheck.Enabled {
		setUpHealthCheckService(d.healthCheck.Port, d.healthCheck.Endpoint)
	}
	log.Fatal(http.ListenAndServe(d.proxy.Listen, d))
	return
}

func (d *daemon) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Header.Get("X-Request-Id") == "" {
		req.Header.Set("X-Request-Id", uuid.New())
	}
	d.handler.ServeHTTP(rw, req)
}

func (d *daemon) LoadConfig(newConfig config.Config) error {
	// We don't support changing the Proxy listen port because it would mean closing and
	// restarting the server. If users want to do this they should restart the process.
	if d.proxy.Listen != "" && d.proxy.Listen != newConfig.Proxy.Listen {
		return fmt.Errorf("SPHINX_LOAD_CONFIG_FAILED. Can't change listen port")
	}

	d.proxy = newConfig.Proxy
	d.healthCheck = newConfig.HealthCheck
	target, _ := url.Parse(d.proxy.Host) // already tested for invalid Host
	proxy := httputil.NewSingleHostReverseProxy(target)
	rateLimiter, err := ratelimiter.New(newConfig)
	if err != nil {
		return fmt.Errorf("SPHINX_LOAD_CONFIG_FAILED: %s", err.Error())
	}

	// Set the proxy and handler daemon fields
	var handler http.Handler
	switch d.proxy.Handler {
	case "http":
		handler = handlers.NewHTTPLimiter(rateLimiter, proxy, d.proxy.AllowOnError)
	case "httplogger":
		handler = handlers.NewHTTPLogger(rateLimiter, proxy)
	default:
		return fmt.Errorf("unrecognized handler %s", d.proxy.Handler)
	}

	d.handler = middleware.New(handler, common.Log, func(req *http.Request) map[string]interface{} {
		return map[string]interface{}{"guid": req.Header.Get("X-Request-Id")}
	})
	return nil
}

// New takes in config.Configuration and creates a sphinx listener
func New(config config.Config) (Daemon, error) {
	out := &daemon{}
	if err := out.LoadConfig(config); err != nil {
		return nil, err
	}
	return out, nil
}

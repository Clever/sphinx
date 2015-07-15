package daemon

import (
	"fmt"
	"github.com/Clever/sphinx/config"
	"github.com/Clever/sphinx/handlers"
	"github.com/Clever/sphinx/ratelimiter"
	"gopkg.in/Clever/kayvee-go.v2"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

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
		log.Printf(kayvee.Format(map[string]interface{}{
			"source": "sphinx",
			"title":  "Health-Check",
			"type":   "counter",
			"value":  1}))
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
	switch d.proxy.Handler {
	case "http":
		d.handler = handlers.NewHTTPLimiter(rateLimiter, proxy, d.proxy.AllowOnError)
		return nil
	case "httplogger":
		d.handler = handlers.NewHTTPLogger(rateLimiter, proxy)
		return nil
	default:
		return fmt.Errorf("unrecognized handler %s", d.proxy.Handler)
	}

}

// NewDaemon takes in config.Configuration and creates a sphinx listener
func New(config config.Config) (Daemon, error) {

	out := &daemon{}
	err := out.LoadConfig(config)
	if err != nil {
		return nil, err
	}

	return out, nil
}

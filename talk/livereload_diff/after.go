package daemon

import (
	"fmt"
	"github.com/Clever/sphinx/config"
	"github.com/Clever/sphinx/handlers"
	"github.com/Clever/sphinx/ratelimiter"
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
	handler http.Handler
	proxy   config.Proxy
}

func (d *daemon) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	d.handler.ServeHTTP(rw, req)
}

func (d *daemon) Start() {
	log.Printf("Listening on %s", d.proxy.Listen)
	log.Fatal(http.ListenAndServe(d.proxy.Listen, d))
	return
}

func (d *daemon) LoadConfig(newConfig config.Config) error {
	// We don't support changing the Proxy listen port because it would mean closing and
	// restarting the server. If users want to do this they should restart the process.
	if d.proxy.Listen != "" && d.proxy.Listen != newConfig.Proxy.Listen {
		return fmt.Errorf("SPHINX_LOAD_CONFIG_FAILED. Can't change listen port")
	}

	d.proxy = newConfig.Proxy
	target, _ := url.Parse(d.proxy.Host) // already tested for invalid Host
	proxy := httputil.NewSingleHostReverseProxy(target)
	rateLimiter, err := ratelimiter.New(newConfig)
	if err != nil {
		return fmt.Errorf("SPHINX_LOAD_CONFIG_FAILED: %s", err.Error())
	}

	switch d.proxy.Handler {
	case "http":
		d.handler = handlers.NewHTTPLimiter(rateLimiter, proxy)
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

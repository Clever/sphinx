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
	ReloadConfig(config config.Config) error
}

type daemon struct {
	rateLimiter ratelimiter.RateLimiter
	handler     handlers.SphinxHandler
	proxy       config.Proxy
}

func (d *daemon) Start() {
	log.Printf("Listening on %s", d.proxy.Listen)
	log.Fatal(http.ListenAndServe(d.proxy.Listen, d.handler))
	return
}

func (d *daemon) ReloadConfig(config config.Config) error {
	rateLimiter, err := ratelimiter.New(config)
	if err != nil {
		return fmt.Errorf("SPHINX_RELOAD_CONFIG_FAILED: %s", err.Error())
	}
	d.rateLimiter = rateLimiter
	d.handler.SetRateLimiter(rateLimiter)
	return nil
}

// NewDaemon takes in config.Configuration and creates a sphinx listener
func New(config config.Config) (Daemon, error) {

	rateLimiter, err := ratelimiter.New(config)
	if err != nil {
		return &daemon{}, fmt.Errorf("SPHINX_INIT_FAILED: %s", err.Error())
	}

	target, _ := url.Parse(config.Proxy.Host) // already tested for invalid Host
	proxy := httputil.NewSingleHostReverseProxy(target)

	out := &daemon{
		proxy:       config.Proxy,
		rateLimiter: rateLimiter,
	}
	switch config.Proxy.Handler {
	case "http":
		out.handler = handlers.NewHTTPLimiter(rateLimiter, proxy)
	case "httplogger":
		out.handler = handlers.NewHTTPLogger(rateLimiter, proxy)
	default:
		return &daemon{}, fmt.Errorf("unrecognized handler %s", config.Proxy.Handler)
	}

	return out, nil
}

package daemon

import (
	"fmt"
	"github.com/Clever/sphinx/config"
	"github.com/Clever/sphinx/handlers"
	"github.com/Clever/sphinx/ratelimit"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Daemon interface {
	Start()
}

type daemon struct {
	config      config.Configuration
	rateLimiter ratelimit.RateLimiter
	proxy       httputil.ReverseProxy
	handler     http.Handler
}

func (d *daemon) Start() {
	log.Printf("Listening on %s", d.config.Proxy().Listen)
	log.Fatal(http.ListenAndServe(d.config.Proxy().Listen, d.handler))
	return
}

// NewDaemon takes in config.Configuration and creates a sphinx listener
func NewDaemon(config config.Configuration) (Daemon, error) {

	rateLimiter, err := ratelimit.NewRateLimiter(config)
	if err != nil {
		return &daemon{}, fmt.Errorf("SPHINX_INIT_FAILED: %s", err.Error())
	}

	target, _ := url.Parse(config.Proxy().Host) // already tested for invalid Host
	proxy := httputil.NewSingleHostReverseProxy(target)

	out := &daemon{
		config:      config,
		rateLimiter: rateLimiter,
	}
	switch config.Proxy().Handler {
	case "http":
		out.handler = handlers.NewHTTPLimiter(rateLimiter, proxy)
	case "httplogger":
		out.handler = handlers.NewHTTPLogger(rateLimiter, proxy)
	default:
		return &daemon{}, fmt.Errorf("unrecognized handler %s", config.Proxy().Handler)
	}

	return out, nil
}

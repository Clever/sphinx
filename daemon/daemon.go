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
	rateLimiter ratelimit.RateLimiter
	handler     http.Handler
	proxy       config.Proxy
}

func (d *daemon) Start() {
	log.Printf("Listening on %s", d.proxy.Listen)
	log.Fatal(http.ListenAndServe(d.proxy.Listen, d.handler))
	return
}

// NewDaemon takes in config.Configuration and creates a sphinx listener
func New(config config.Config) (Daemon, error) {

	rateLimiter, err := ratelimit.New(config)
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

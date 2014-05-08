package main

import (
	"flag"
	"fmt"
	"github.com/Clever/sphinx"
	"github.com/Clever/sphinx/handlers"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	configfile = flag.String("config", "example.yaml", "/path/to/configuration.yaml")
	validate   = flag.Bool("validate", false, "Validate configuration and exit")
)

type daemon struct {
	config      sphinx.Configuration
	rateLimiter sphinx.RateLimiter
	proxy       httputil.ReverseProxy
	handler     http.Handler
}

func (d *daemon) Start() {
	log.Printf("Listening on %s", d.config.Proxy.Listen)
	log.Fatal(http.ListenAndServe(d.config.Proxy.Listen, d.handler))
	return
}

// NewDaemon takes in sphinx.Configuration and creates a sphinx listener
func NewDaemon(config sphinx.Configuration) (daemon, error) {

	rateLimiter, err := sphinx.NewRateLimiter(config)
	if err != nil {
		return daemon{}, fmt.Errorf("SPHINX_INIT_FAILED: %s", err.Error())
	}

	target, _ := url.Parse(config.Proxy.Host) // already tested for invalid Host
	proxy := httputil.NewSingleHostReverseProxy(target)

	out := daemon{
		config:      config,
		rateLimiter: rateLimiter,
	}
	switch config.Proxy.Handler {
	case "http":
		out.handler = handlers.NewHTTPLimiter(rateLimiter, proxy)
	case "httplogger":
		out.handler = handlers.NewHTTPLogger(rateLimiter, proxy)
	default:
		return daemon{}, fmt.Errorf("unrecognized handler %s", config.Proxy.Handler)
	}

	return out, nil
}

func main() {

	flag.Parse()

	config, err := sphinx.NewConfiguration(*configfile)
	if err != nil {
		log.Fatalf("LOAD_CONFIG_FAILED: %s", err.Error())
	}

	sphinxd, err := NewDaemon(config)
	if err != nil {
		log.Fatal(err)
	}

	if *validate {
		print("configuration parsed and Sphinx loaded fine. not starting dameon.")
		return
	}

	sphinxd.Start()
}

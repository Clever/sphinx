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
	//"os"
	//"os/signal"
	//"syscall"
)

var (
	configfile = flag.String("config", "example.yaml", "/path/to/configuration.yaml")
	validate   = flag.Bool("validate", false, "Validate configuration and exit")
)

type Daemon struct {
	config      sphinx.Configuration
	ratelimiter sphinx.RateLimiter
	proxy       httputil.ReverseProxy
	handler     http.Handler
}

//func (d *Daemon) Reload(config Configuration) bool {
//}

func (d *Daemon) Start() {
	log.Printf("Listening on %s", d.config.Proxy.Listen)
	log.Fatal(http.ListenAndServe(d.config.Proxy.Listen, d.handler))
	return
}

//func (d *Daemon) Quit() bool {

//}

func NewDaemon(config sphinx.Configuration) (Daemon, error) {

	ratelimiter, err := sphinx.NewRateLimiter(config)
	if err != nil {
		return Daemon{}, fmt.Errorf("SPHINX_INIT_FAILED: %s", err.Error())
	}

	target, _ := url.Parse(config.Proxy.Host) // already tested for invalid Host
	proxy := httputil.NewSingleHostReverseProxy(target)

	var httplimiter http.Handler
	switch config.Proxy.Handler {
	case "http":
		httplimiter = handlers.NewHTTPLogger(ratelimiter, proxy)
	case "httplogger":
		httplimiter = handlers.NewHTTPLimiter(ratelimiter, proxy)
	default:
		return Daemon{}, fmt.Errorf("Sphinx only supports the http handler")
	}

	return Daemon{
		config:      config,
		ratelimiter: ratelimiter,
		handler:     httplimiter,
	}, nil

}

func main() {

	flag.Parse()

	config, err := sphinx.NewConfiguration(*configfile)
	if err != nil {
		log.Fatalf("LOAD_CONFIG_FAILED: %s", err.Error())
	}

	daemon, err := NewDaemon(config)
	if err != nil {
		log.Fatal(err)
	}

	if *validate {
		print("Configuration parsed and Sphinx loaded fine. not starting dameon.")
		return
	}

	daemon.Start()
}

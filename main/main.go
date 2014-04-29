package main

import (
	"flag"
	"github.com/Clever/sphinx"
	"github.com/Clever/sphinx/handlers"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
)

var (
	configfile = flag.String("config", "example.yaml", "/path/to/configuration.yaml")
	validate   = flag.Bool("validate", false, "Validate configuration and exit")
)

type Daemon struct {
}

func (d *Daemon) Reload(config Configuration) bool {

}

func (d *Daemon) Quit() bool {

}

func main() {

	flag.Parse()
	config, err := sphinx.NewConfiguration(*configfile)
	if err != nil {
		log.Fatalf("LOAD_CONFIG_FAILED: %s", err.Error())
	}
	ratelimiter, err := sphinx.NewRateLimiter(config)
	if err != nil {
		log.Fatalf("SPHINX_INIT_FAILED: %s", err.Error())
	}

	// if configuration says that use http
	if config.Proxy.Handler != "http" {
		log.Fatalf("Sphinx only supports the http handler")
	}

	target, _ := url.Parse(config.Proxy.Host)
	proxy := httputil.NewSingleHostReverseProxy(target)
	httplimiter := handlers.NewHTTPLogger(ratelimiter, proxy)

	if *validate {
		print("Configuration parsed and Sphinx loaded fine. not starting dameon.")
		return
	}

	log.Printf("Listening on %s", config.Proxy.Listen)
	log.Fatal(http.ListenAndServe(config.Proxy.Listen, wrapper))
}

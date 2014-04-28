package main

import (
	"flag"
	"github.com/Clever/sphinx"
	handlers "github.com/Clever/sphinx/handlers"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	configfile = flag.String("config", "example.yaml", "/path/to/configuration.yaml")
	validate   = flag.Bool("validate", false, "Validate configuration and exit")
)

func main() {

	flag.Parse()
	config, _ := sphinx.NewConfiguration(*configfile)
	ratelimiter, _ := sphinx.NewRateLimiter(config)

	// if configuration says that use http
	if config.Forward.Scheme != "http" {
		return
	}

	target, _ := url.Parse(config.Forward.Host)
	proxy := httputil.NewSingleHostReverseProxy(target)
	httplimiter := handlers.NewHTTPLogger(ratelimiter, proxy)

	if *validate {
		print("configuration is fine. not starting dameon.")
		return
	}

	println("listening on 8080")
	log.Fatal(http.ListenAndServe(":8080", httplimiter))
}

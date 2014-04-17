package main

import (
	"flag"
	"github.com/Clever/sphinx"
	handlers "github.com/Clever/sphinx/handlers"
	"net/http/httputil"
	"net/url"
)

var (
	configfile  = flag.String("config", "example.yaml", "/path/to/configuration.yaml")
	checkconfig = flag.Bool("check-config", false, "Validate configuration and exit")
)

func main() {

	flag.Parse()
	config, err := sphinx.NewConfiguration(*configfile)
	rateLimiter, err := sphinx.NewRateLimiter(config)

	// if configuration says that use http
	if config.Forward.Scheme == "http" {
		target, _ := url.Parse(config.Forward.Host)
		proxy := httputil.NewSingleHostReverseProxy(target)
		_ = handlers.HTTPRateLimiter{
		//rateLimiter,
		//proxy,
		}

		if !*checkconfig {
			print("configuration is fine. not starting dameon.")
			return
		}

		print(proxy, rateLimiter, err)

		//http.ListenAndServe(config.Forward.Listen, http.Handler{})
	}
}

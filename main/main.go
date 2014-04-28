package main

import (
	"flag"
	"github.com/Clever/sphinx"
	"net/http/httputil"
	"net/url"
)

var (
	configfile = flag.String("config", "example.yaml", "/path/to/configuration.yaml")
	validate   = flag.Bool("validate", false, "Validate configuration and exit")
)

func main() {

	flag.Parse()
	config, err := sphinx.NewConfiguration(*configfile)
	rateLimiter, err := sphinx.NewRateLimiter(config)

	// if configuration says that use http
	if config.Forward.Scheme == "http" {
		target, _ := url.Parse(config.Forward.Host)
		proxy := httputil.NewSingleHostReverseProxy(target)
		//_ = handlers.HTTPRateLimiter{
		//rateLimiter,
		//proxy,
		//}

		if !*validate {
			print("configuration is fine. not starting dameon.")
			return
		}

		print(proxy, rateLimiter, err)

		//http.ListenAndServe(config.Forward.Listen, http.Handler{})
	}
}

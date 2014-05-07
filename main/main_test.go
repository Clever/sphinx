package main

import (
	"github.com/Clever/sphinx"
	"github.com/Clever/sphinx/handlers"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"testing"
)

var host = "http://localhost:8081"

type Handler struct{}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte{})
}

func setUpLocalServer() {
	go http.ListenAndServe(":8081", Handler{})
}

func setUpHTTPLimiter(b *testing.B) {
	config, err := sphinx.NewConfiguration("../example.yaml")
	if err != nil {
		b.Fatalf("LOAD_CONFIG_FAILED: %s", err.Error())
	}
	rateLimiter, err := sphinx.NewRateLimiter(config)
	if err != nil {
		b.Fatalf("SPHINX_INIT_FAILED: %s", err.Error())
	}

	// if configuration says that use http
	if config.Proxy.Handler != "http" {
		b.Fatalf("sphinx only supports the http handler")
	}

	// ignore the url in the config and use localhost
	target, _ := url.Parse(host)
	proxy := httputil.NewSingleHostReverseProxy(target)
	httpLimiter := handlers.NewHTTPLimiter(rateLimiter, proxy)

	config.Proxy.Listen = ":8082"
	go http.ListenAndServe(config.Proxy.Listen, httpLimiter)
}

func makeRequestTo(port string) error {
	// Add basic auth so that we match some buckets.
	if resp, err := http.Get("http://user:pass@localhost" + port); err != nil {
		log.Printf("got resp %#v", resp)
		return err
	}
	return nil
}

func BenchmarkNoLimiter(b *testing.B) {
	setUpLocalServer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := makeRequestTo(":8081"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReasonableConfig(b *testing.B) {
	setUpLocalServer()
	setUpHTTPLimiter(b)
	// So we don't spam with logs
	log.SetOutput(ioutil.Discard)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		makeRequestTo(":8082")
	}
}

package daemon

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Clever/kayvee-go/v7/logger"
	"github.com/Clever/kayvee-go/v7/middleware"
	"github.com/Clever/sphinx/common"
	"github.com/Clever/sphinx/config"
	"github.com/Clever/sphinx/handlers"
	"github.com/Clever/sphinx/ratelimiter"
	"github.com/pborman/uuid"
	"gopkg.in/tylerb/graceful.v1"
)

// Daemon represents a daemon server
type Daemon interface {
	Start()
	LoadConfig(config config.Config) error
}

type daemon struct {
	handler     http.Handler
	proxy       config.Proxy
	healthCheck config.HealthCheck
}

// setUpHealthCheckService sets up a health check service at the given port
// that can be pinged at the given endpoint to determine if Sphinx is still
// running.
func setUpHealthCheckService(port string, endpoint string) {
	mux := http.NewServeMux()
	mux.HandleFunc(endpoint, func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		common.Log.Counter("Health-Check")
	})
	go http.ListenAndServe(":"+port, mux)
	log.Printf("Health-check listening on:%s%s", port, endpoint)
}

func (d *daemon) Start() {
	log.Printf("Limiter listening on %s", d.proxy.Listen)
	// Only set up the health check service if it is enabled.
	if d.healthCheck.Enabled {
		setUpHealthCheckService(d.healthCheck.Port, d.healthCheck.Endpoint)
	}
	if err := graceful.RunWithErr(d.proxy.Listen, 30*time.Second, d); err != nil {
		log.Fatal(err)
	}
	return
}

func (d *daemon) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Header.Get("X-Request-Id") == "" {
		req.Header.Set("X-Request-Id", uuid.New())
	}
	d.handler.ServeHTTP(rw, req)
}

func (d *daemon) LoadConfig(newConfig config.Config) error {
	// We don't support changing the Proxy listen port because it would mean closing and
	// restarting the server. If users want to do this they should restart the process.
	if d.proxy.Listen != "" && d.proxy.Listen != newConfig.Proxy.Listen {
		return fmt.Errorf("SPHINX_LOAD_CONFIG_FAILED. Can't change listen port")
	}

	d.proxy = newConfig.Proxy
	d.healthCheck = newConfig.HealthCheck
	target, _ := url.Parse(d.proxy.Host) // already tested for invalid Host

	// This is same as httputil.NewSingleHostReverseProxy with a change of adding "req.Host = targetURL.Host" https://github.com/golang/go/issues/28168
	// In future version of go we can use ReverseProxy.Rewrite instead https://github.com/golang/go/commit/a55793835f16d0242be18aff4ec0bd13494175bd
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path, req.URL.RawPath = joinURLPath(target, req.URL)
		req.Host = target.Host
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}

	proxy := &httputil.ReverseProxy{Director: director}

	// An attempt to fix cancelled DNS requests resulting in 502 errors
	// See https://groups.google.com/d/msg/golang-nuts/oiBBZfUb2hM/9S_JB6g2EAAJ
	proxy.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).Dial,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	rateLimiter, err := ratelimiter.New(newConfig)
	if err != nil {
		return fmt.Errorf("SPHINX_LOAD_CONFIG_FAILED: %s", err.Error())
	}

	// Set the proxy and handler daemon fields
	var handler http.Handler
	switch d.proxy.Handler {
	case "http":
		handler = handlers.NewHTTPLimiter(rateLimiter, proxy, d.proxy.AllowOnError)
	case "httplogger":
		handler = handlers.NewHTTPLogger(rateLimiter, proxy)
	default:
		return fmt.Errorf("unrecognized handler %s", d.proxy.Handler)
	}

	var pathRegex = regexp.MustCompile(`(/.*)(/[0-9a-f]{24})(.*)`)

	middleware.EnableRollups(context.Background(), logger.New("sphinx"), 20*time.Second)
	d.handler = middleware.New(handler, "sphinx", func(req *http.Request) map[string]interface{} {
		path := req.URL.Path
		// matches will be empty if the path has no 24-character hex ID
		// otherwise it will have one element, namely an array with
		// - the whole match (i.e. the whole path)
		// - capture group 1 (everything before the ID)
		// - capture group 2 (the ID)
		// - capture group 3 (everything after the ID)
		matches := pathRegex.FindAllStringSubmatch(path, 1)
		if len(matches) == 1 && len(matches[0]) == 4 {
			path = matches[0][1] + matches[0][3]
		}
		return map[string]interface{}{
			"guid": req.Header.Get("X-Request-Id"),
			"op":   path, // add op key since log rollups are keyed on method, status, and op
		}
	})
	return nil
}

// New takes in config.Configuration and creates a sphinx listener
func New(config config.Config) (Daemon, error) {
	out := &daemon{}
	if err := out.LoadConfig(config); err != nil {
		return nil, err
	}
	return out, nil
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}

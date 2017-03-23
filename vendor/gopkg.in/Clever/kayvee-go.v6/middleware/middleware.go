/*
Package middleware provides a customizable Kayvee logging middleware for HTTP servers.

	logHandler := New(myHandler, myLogger, func(req *http.Request) map[string]interface{} {
		// Add Gorilla mux vars to the log, just because
		return mux.Vars(req)
	})

*/
package middleware

import (
	"net/http"
	"os"
	"time"

	"gopkg.in/Clever/kayvee-go.v6/logger"
)

var defaultHandler = func(req *http.Request) map[string]interface{} {
	return map[string]interface{}{
		"method": req.Method,
		"path":   req.URL.Path,
		"params": req.URL.RawQuery,
		"ip":     getIP(req),
	}
}

type logHandler struct {
	handlers []func(req *http.Request) map[string]interface{}
	h        http.Handler
	isCanary bool
	source   string
}

func (l *logHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	start := time.Now()

	// create and inject a logger into req.Context
	lggr := logger.New(l.source)
	req = req.WithContext(logger.NewContext(req.Context(), lggr))

	lrw := &loggedResponseWriter{
		status:         200,
		ResponseWriter: w,
		length:         0,
	}
	l.h.ServeHTTP(lrw, req)
	duration := time.Since(start)

	data := l.applyHandlers(req, map[string]interface{}{
		"response-time": duration,
		"response-size": lrw.length,
		"status-code":   lrw.status,
		"via":           "kayvee-middleware",
		"canary":        l.isCanary,
	})

	switch logLevelFromStatus(lrw.status) {
	case logger.Error:
		lggr.ErrorD("request-finished", data)
	default:
		lggr.InfoD("request-finished", data)
	}
}

func (l *logHandler) applyHandlers(req *http.Request, finalizer map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	writeData := func(data map[string]interface{}) {
		for key, val := range data {
			result[key] = val
		}
	}

	for _, handler := range l.handlers {
		writeData(handler(req))
	}
	// Write reserved fields last to make sure nothing overwrites them
	writeData(defaultHandler(req))
	writeData(finalizer)

	return result
}

// New takes in an http Handler to wrap with logging, the logger source name to use, and any amount of
// optional handlers to customize the data that's logged.
// On every request, the middleware will create a logger and place it in req.Context().
func New(h http.Handler, source string, handlers ...func(*http.Request) map[string]interface{}) http.Handler {
	isCanary := false

	canaryFlag := os.Getenv("_CANARY")
	if canaryFlag == "1" {
		isCanary = true
	}
	return &logHandler{
		handlers: handlers,
		h:        h,
		isCanary: isCanary,
		source:   source,
	}
}

// HeaderHandler takes in any amount of headers and returns a handler that adds those headers.
func HeaderHandler(headers ...string) func(*http.Request) map[string]interface{} {
	return func(req *http.Request) map[string]interface{} {
		result := map[string]interface{}{}
		for _, header := range headers {
			if val := req.Header.Get(header); val != "" {
				result[header] = val
			}
		}
		return result
	}
}

type loggedResponseWriter struct {
	status int
	http.ResponseWriter
	length int
}

func (w *loggedResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *loggedResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	return n, err
}

func getIP(req *http.Request) string {
	forwarded := req.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}
	return req.RemoteAddr
}

func logLevelFromStatus(status int) logger.LogLevel {
	if status >= 499 {
		return logger.Error
	}
	return logger.Info
}

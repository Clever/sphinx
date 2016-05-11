/*
Package middleware provides a customizable Kayvee logging middleware for HTTP servers.

	logHandler := New(myHandler, myLogger, func(req *http.Request) map[string]interface{} {
		// Add Gorilla mux vars to the log, just because
		return mux.Vays(req)
	})

*/
package middleware

import (
	"net/http"
	"time"

	"gopkg.in/Clever/kayvee-go.v3/logger"
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
	logger   *logger.Logger
}

func (l *logHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	start := time.Now()

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
	})

	switch logLevelFromStatus(lrw.status) {
	case logger.Error:
		l.logger.ErrorD("request-finished", data)
	default:
		l.logger.InfoD("request-finished", data)
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

// New takes in an http Handler to wrap with logging, the logger to use, and any amount of
// optional handlers to customize the data that's logged.
func New(h http.Handler, logger *logger.Logger, handlers ...func(*http.Request) map[string]interface{}) http.Handler {
	return &logHandler{
		logger:   logger,
		handlers: handlers,
		h:        h,
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

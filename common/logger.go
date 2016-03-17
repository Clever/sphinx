package common

import (
	"gopkg.in/Clever/kayvee-go.v3/logger"

	"net/http"
	"strings"
)

// Log is a Kayvee.Logger singleton to be used in Sphinx
var Log *logger.Logger

func init() {
	Log = logger.New("sphinx")
}

// M is an alias for map[String]interface{} to make log lines less painful to write.
type M map[string]interface{}

// ConcatWithRequest concats the request to a given map[String]interface{} for use with Kayvee
func ConcatWithRequest(data M, req Request) M {
	var kvData = M{
		"path":       req["path"],
		"remoteaddr": req["remoteaddr"],
		"method":     req["method"],
	}
	for header, values := range req["headers"].(http.Header) {
		kvData[header] = strings.Join(values, ";")
	}
	for key, val := range data {
		kvData[key] = val
	}

	return kvData
}

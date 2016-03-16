package common

import (
	"gopkg.in/Clever/kayvee-go.v2/logger"

	"net/http"
	"strings"
)

var Log *logger.Logger

func init() {
	Log = logger.New("sphinx")
}

// M is an alias for map[Stringsing]interface{} to make log lines less painful to write.
type M map[string]interface{}

func LogWithRequest(data M, req Request) M {
	var kvData = M{
		"path":       req["path"],
		"remoteaddr": req["remoteaddr"],
	}
	for header, values := range req["headers"].(http.Header) {
		kvData[header] = strings.Join(values, ";")
	}
	for key, val := range data {
		kvData[key] = val
	}

	return kvData
}

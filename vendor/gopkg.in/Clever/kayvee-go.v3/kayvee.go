package kayvee

import (
	"encoding/json"
	"log"
)

// Log Levels:

// LogLevel denotes the level of a logging
type LogLevel string

// Constants used to define different LogLevels supported
const (
	Unknown  LogLevel = "unknown"
	Critical          = "critical"
	Error             = "error"
	Warning           = "warning"
	Info              = "info"
	Trace             = "trace"
)

// Internal defaults used by Logger.
const (
	defaultFlags = log.LstdFlags | log.Lshortfile
)

// Format converts a map to a string of space-delimited key=val pairs
func Format(data map[string]interface{}) string {
	formattedString, _ := json.Marshal(data)
	return string(formattedString)
}

// FormatLog is similar to Format, but takes additional reserved params to promote logging best-practices
func FormatLog(source string, level LogLevel, title string, data map[string]interface{}) string {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["source"] = source
	data["level"] = level
	data["title"] = title
	return Format(data)
}

// Logger is an interface satisfied by all loggers that use kayvee to Log results
type Logger interface {
	Info(title string, data map[string]interface{})
	Warning(title string, data map[string]interface{})
	Error(title string, data map[string]interface{}, err error)
}

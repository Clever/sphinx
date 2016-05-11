package logger

import (
	"io"
	"log"
	"os"
	"strings"

	kv "gopkg.in/Clever/kayvee-go.v3"
)

/////////////////////
//
//  Helper definitions
//
/////////////////////

// Formatter is a function type that takes a map and returns a formatted string with the contents of the map
type Formatter func(data map[string]interface{}) string

// LogLevel is an enum is used to denote level of logging
type LogLevel int

// M is a convenience type for passing data into a log message.
type M map[string]interface{}

// Constants used to define different LogLevels supported
const (
	Debug LogLevel = iota
	Info
	Warning
	Error
	Critical
)

var reservedKeyNames = map[string]bool{
	"title":  true,
	"source": true,
	"value":  true,
	"type":   true,
}

var logLevelNames = map[LogLevel]string{
	Debug:    "debug",
	Info:     "info",
	Warning:  "warning",
	Error:    "error",
	Critical: "critical",
}

func (l LogLevel) String() string {
	switch l {
	case Debug:
		fallthrough
	case Info:
		fallthrough
	case Warning:
		fallthrough
	case Error:
		fallthrough
	case Critical:
		return logLevelNames[l]
	}
	return ""
}

/////////////////////////////
//
//	Logger
//
/////////////////////////////

// Logger provides customization of log messages. We can change globals, default log level, formatting, and output destination.
type Logger struct {
	globals   map[string]interface{}
	logLvl    LogLevel
	formatter Formatter
	logWriter *log.Logger
}

// SetConfig allows configuration changes in one go
func (l *Logger) SetConfig(source string, logLvl LogLevel, formatter Formatter, output io.Writer) {
	if l.globals == nil {
		l.globals = make(map[string]interface{})
	}
	l.globals["source"] = source
	l.logLvl = logLvl
	l.formatter = formatter
	l.logWriter = log.New(output, "", 0) // No prefixes
}

// SetLogLevel sets the default log level threshold
func (l *Logger) SetLogLevel(logLvl LogLevel) {
	l.logLvl = logLvl
}

// SetFormatter sets the formatter function to use
func (l *Logger) SetFormatter(formatter Formatter) {
	l.formatter = formatter
}

// SetOutput changes the output destination of the logger
func (l *Logger) SetOutput(output io.Writer) {
	l.logWriter = log.New(output, "", 0) // No prefixes
}

// Logging functions

// Debug takes a string and logs with LogLevel = Debug
func (l *Logger) Debug(title string) {
	l.DebugD(title, M{})
}

// Info takes a string and logs with LogLevel = Info
func (l *Logger) Info(title string) {
	l.InfoD(title, M{})
}

// Warn takes a string and logs with LogLevel = Warning
func (l *Logger) Warn(title string) {
	l.WarnD(title, M{})
}

// Error takes a string and logs with LogLevel = Error
func (l *Logger) Error(title string) {
	l.ErrorD(title, M{})
}

// Critical takes a string and logs with LogLevel = Critical
func (l *Logger) Critical(title string) {
	l.CriticalD(title, M{})
}

// Counter takes a string and logs with LogLevel = Info, type = counter, and value = 1
func (l *Logger) Counter(title string) {
	l.CounterD(title, 1, M{}) // Defaults to a value of 1
}

// GaugeInt takes a string and integer value. It logs with LogLevel = Info, type = gauge, and value = value
func (l *Logger) GaugeInt(title string, value int) {
	l.GaugeIntD(title, value, M{})
}

// GaugeFloat takes a string and float value. It logs with LogLevel = Info, type = gauge, and value = value
func (l *Logger) GaugeFloat(title string, value float64) {
	l.GaugeFloatD(title, value, M{})
}

// DebugD takes a string and data map. It logs with LogLevel = Debug
func (l *Logger) DebugD(title string, data map[string]interface{}) {
	data["title"] = title
	l.logWithLevel(Debug, data)
}

// InfoD takes a string and data map. It logs with LogLevel = Info
func (l *Logger) InfoD(title string, data map[string]interface{}) {
	data["title"] = title
	l.logWithLevel(Info, data)
}

// WarnD takes a string and data map. It logs with LogLevel = Warning
func (l *Logger) WarnD(title string, data map[string]interface{}) {
	data["title"] = title
	l.logWithLevel(Warning, data)
}

// ErrorD takes a string and data map. It logs with LogLevel = Error
func (l *Logger) ErrorD(title string, data map[string]interface{}) {
	data["title"] = title
	l.logWithLevel(Error, data)
}

// CriticalD takes a string and data map. It logs with LogLevel = Critical
func (l *Logger) CriticalD(title string, data map[string]interface{}) {
	data["title"] = title
	l.logWithLevel(Critical, data)
}

// CounterD takes a string, value, and data map. It logs with LogLevel = Info, type = counter, and value = value
func (l *Logger) CounterD(title string, value int, data map[string]interface{}) {
	data["title"] = title
	data["value"] = value
	data["type"] = "counter"
	l.logWithLevel(Info, data)
}

// GaugeIntD takes a string, an integer value, and data map. It logs with LogLevel = Info, type = gauge, and value = value
func (l *Logger) GaugeIntD(title string, value int, data map[string]interface{}) {
	l.gauge(title, value, data)
}

// GaugeFloatD takes a string, a float value, and data map. It logs with LogLevel = Info, type = gauge, and value = value
func (l *Logger) GaugeFloatD(title string, value float64, data map[string]interface{}) {
	l.gauge(title, value, data)
}

func (l *Logger) gauge(title string, value interface{}, data map[string]interface{}) {
	data["title"] = title
	data["value"] = value
	data["type"] = "gauge"
	l.logWithLevel(Info, data)
}

// Actual logging. Handles whether to output based on log level and
// unifies the passed in data with the stored globals
func (l *Logger) logWithLevel(logLvl LogLevel, data map[string]interface{}) {
	if logLvl < l.logLvl {
		// No log output
		return
	}
	data["level"] = logLvl.String()
	for key, value := range l.globals {
		if _, ok := data[key]; ok {
			// Values in the data map override the globals
			continue
		}
		data[key] = value
	}

	logString := l.formatter(data)
	l.logWriter.Println(logString)
}

// New creates a *logger.Logger. Default values are Debug LogLevel, kayvee Formatter, and std.err output.
func New(source string) *Logger {
	return NewWithContext(source, nil)
}

// NewWithContext creates a *logger.Logger. Default values are Debug LogLevel, kayvee Formatter, and std.err output.
func NewWithContext(source string, contextValues map[string]interface{}) *Logger {
	context := M{}
	for k, v := range contextValues {
		if reservedKeyNames[strings.ToLower(k)] {
			log.Printf("WARN: kayvee logger reserves '%s' from being set as context", k)
			continue
		}
		context[k] = v
	}
	if os.Getenv("_DEPLOY_ENV") != "" {
		context["deploy_env"] = os.Getenv("_DEPLOY_ENV")
	}

	logObj := Logger{
		globals: context,
	}

	var logLvl LogLevel
	strLogLvl := os.Getenv("KAYVEE_LOG_LEVEL")
	if strLogLvl == "" {
		logLvl = Debug
	} else {
		for key, val := range logLevelNames {
			if strings.ToLower(strLogLvl) == val {
				logLvl = key
				break
			}
		}
	}
	logObj.SetConfig(source, logLvl, kv.Format, os.Stderr)
	return &logObj
}

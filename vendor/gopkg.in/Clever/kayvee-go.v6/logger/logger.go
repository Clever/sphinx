package logger

import (
	"io"
	"log"
	"os"
	"strings"

	kv "gopkg.in/Clever/kayvee-go.v6"
	"gopkg.in/Clever/kayvee-go.v6/router"
)

/////////////////////
//
//  Helper definitions
//
/////////////////////

// Formatter is a function type that takes a map and returns a formatted string with the contents of the map
type Formatter func(data map[string]interface{}) string

// M is a convenience type for passing data into a log message.
type M map[string]interface{}

// LogLevel is an enum is used to denote level of logging
type LogLevel int

// Constants used to define different LogLevels supported
const (
	Debug LogLevel = iota
	Info
	Warning
	Error
	Critical
)

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

// Logger is the default implementation of KayveeLogger.
// It provides customization of globals, default log level, formatting, and output destination.
type Logger struct {
	globals   map[string]interface{}
	logLvl    LogLevel
	fLogger   formatLogger
	logRouter router.Router
}

var globalRouter router.Router

var reservedKeyNames = map[string]bool{
	"title":   true,
	"source":  true,
	"value":   true,
	"type":    true,
	"level":   true,
	"_kvmeta": true,
}

// SetGlobalRouting installs a new log router onto the KayveeLogger with the
// configuration specified in `filename`. For convenience, the KayveeLogger is expected
// to return itself as the first return value.
func SetGlobalRouting(filename string) error {
	var err error
	globalRouter, err = router.NewFromConfig(filename)
	return err
}

// SetConfig implements the method for the KayveeLogger interface.
func (l *Logger) SetConfig(source string, logLvl LogLevel, formatter Formatter, output io.Writer) {
	if l.globals == nil {
		l.globals = make(map[string]interface{})
	}
	l.globals["source"] = source
	l.logLvl = logLvl
	l.fLogger.setFormatter(formatter)
	l.fLogger.setOutput(output)
}

// AddContext implements the method for the KayveeLogger interface.
func (l *Logger) AddContext(key, val string) {
	updateContextMapIfNotReserved(l.globals, key, val)
}

// SetRouter implements the method for the KayveeLogger interface.
func (l *Logger) SetRouter(router router.Router) {
	l.logRouter = router
}

// SetLogLevel implements the method for the KayveeLogger interface.
func (l *Logger) SetLogLevel(logLvl LogLevel) {
	l.logLvl = logLvl
}

// SetFormatter implements the method for the KayveeLogger interface.
func (l *Logger) SetFormatter(formatter Formatter) {
	l.fLogger.setFormatter(formatter)
}

// SetOutput implements the method for the KayveeLogger interface.
func (l *Logger) SetOutput(output io.Writer) {
	l.fLogger.setOutput(output)
}

func (l *Logger) setFormatLogger(fl formatLogger) {
	l.fLogger = fl
}

// Debug implements the method for the KayveeLogger interface.
func (l *Logger) Debug(title string) {
	l.DebugD(title, M{})
}

// Info implements the method for the KayveeLogger interface.
func (l *Logger) Info(title string) {
	l.InfoD(title, M{})
}

// Warn implements the method for the KayveeLogger interface.
func (l *Logger) Warn(title string) {
	l.WarnD(title, M{})
}

// Error implements the method for the KayveeLogger interface.
func (l *Logger) Error(title string) {
	l.ErrorD(title, M{})
}

// Critical implements the method for the KayveeLogger interface.
func (l *Logger) Critical(title string) {
	l.CriticalD(title, M{})
}

// Counter implements the method for the KayveeLogger interface.
// Logs with type = gauge, and value = value
func (l *Logger) Counter(title string) {
	l.CounterD(title, 1, M{}) // Defaults to a value of 1
}

// GaugeInt implements the method for the KayveeLogger interface.
// Logs with type = gauge, and value = value
func (l *Logger) GaugeInt(title string, value int) {
	l.GaugeIntD(title, value, M{})
}

// GaugeFloat implements the method for the KayveeLogger interface.
// Logs with type = gauge, and value = value
func (l *Logger) GaugeFloat(title string, value float64) {
	l.GaugeFloatD(title, value, M{})
}

// DebugD implements the method for the KayveeLogger interface.
func (l *Logger) DebugD(title string, data map[string]interface{}) {
	data["title"] = title
	l.logWithLevel(Debug, data)
}

// InfoD implements the method for the KayveeLogger interface.
func (l *Logger) InfoD(title string, data map[string]interface{}) {
	data["title"] = title
	l.logWithLevel(Info, data)
}

// WarnD implements the method for the KayveeLogger interface.
func (l *Logger) WarnD(title string, data map[string]interface{}) {
	data["title"] = title
	l.logWithLevel(Warning, data)
}

// ErrorD implements the method for the KayveeLogger interface.
func (l *Logger) ErrorD(title string, data map[string]interface{}) {
	data["title"] = title
	l.logWithLevel(Error, data)
}

// CriticalD implements the method for the KayveeLogger interface.
func (l *Logger) CriticalD(title string, data map[string]interface{}) {
	data["title"] = title
	l.logWithLevel(Critical, data)
}

// CounterD implements the method for the KayveeLogger interface.
// Logs with type = gauge, and value = value
func (l *Logger) CounterD(title string, value int, data map[string]interface{}) {
	data["title"] = title
	data["value"] = value
	data["type"] = "counter"
	l.logWithLevel(Info, data)
}

// GaugeIntD implements the method for the KayveeLogger interface.
// Logs with type = gauge, and value = value
func (l *Logger) GaugeIntD(title string, value int, data map[string]interface{}) {
	l.gauge(title, value, data)
}

// GaugeFloatD implements the method for the KayveeLogger interface.
// Logs with type = gauge, and value = value
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
	if l.logRouter != nil {
		data["_kvmeta"] = l.logRouter.Route(data)
	} else if globalRouter != nil {
		data["_kvmeta"] = globalRouter.Route(data)
	}

	l.fLogger.formatAndLog(data)
}

// updateContextMapIfNotReserved updates context[key] to val if key is not in the reserved list.
func updateContextMapIfNotReserved(context M, key string, val interface{}) {
	if reservedKeyNames[strings.ToLower(key)] {
		log.Printf("WARN: kayvee logger reserves '%s' from being set as context", key)
		return
	}
	context[key] = val
}

// New creates a *logger.Logger. Default values are Debug LogLevel, kayvee Formatter, and std.err output.
func New(source string) KayveeLogger {
	return NewWithContext(source, nil)
}

// NewWithContext creates a *logger.Logger. Default values are Debug LogLevel, kayvee Formatter, and std.err output.
func NewWithContext(source string, contextValues map[string]interface{}) KayveeLogger {
	context := M{}
	for k, v := range contextValues {
		updateContextMapIfNotReserved(context, k, v)
	}
	if os.Getenv("_DEPLOY_ENV") != "" {
		context["deploy_env"] = os.Getenv("_DEPLOY_ENV")
	}

	logObj := Logger{
		globals: context,
	}

	fl := defaultFormatLogger{}
	logObj.fLogger = &fl

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

/////////////////////////////
//
//	FormatLogger
//
/////////////////////////////

// formatLogger is an interface that abstracts the last steps in submitting a log
// message to a Logger: formatting and log writing. It can be replaced for testing.
// This is not yet exported, but could be if clients want customization of the
// format and writing steps.
type formatLogger interface {
	// formatAndLog processes the given data map into a log line and writes it
	formatAndLog(data map[string]interface{})

	// setFormatter specifies the Formatter function to use in formatAndLog
	setFormatter(formatter Formatter)

	// setOutput specifies the output destination to use in formatAndLog
	setOutput(output io.Writer)
}

// defaultFormatLogger provides default implementation of a formatLogger.
type defaultFormatLogger struct {
	formatter Formatter
	logWriter *log.Logger
}

// formatAndLog implements the formatLogger interface for *defaultFormatLogger.
func (fl *defaultFormatLogger) formatAndLog(data map[string]interface{}) {
	logString := fl.formatter(data)
	fl.logWriter.Println(logString)
}

// setFormat implements the formatLogger interface for *defaultFormatLogger.
func (fl *defaultFormatLogger) setFormatter(formatter Formatter) {
	fl.formatter = formatter
}

// setOutput implements the formatLogger interface for *defaultFormatLogger.
func (fl *defaultFormatLogger) setOutput(output io.Writer) {
	fl.logWriter = log.New(output, "", 0) // No prefixes
}

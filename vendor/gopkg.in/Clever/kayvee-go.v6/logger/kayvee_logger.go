package logger

import (
	"io"

	"gopkg.in/Clever/kayvee-go.v6/router"
)

/////////////////////////////
//
//	KayveeLogger interface
//
/////////////////////////////

// KayveeLogger is the main logging interface, providing customization of log messages.
type KayveeLogger interface {

	//
	// Configuration
	//

	// AddContext adds a new key-val to be logged with all log messages.
	AddContext(key, val string)

	// SetConfig allows configuration changes in one go
	SetConfig(source string, logLvl LogLevel, formatter Formatter, output io.Writer)

	// SetFormatter sets the formatter function to use
	SetFormatter(formatter Formatter)

	// SetLogLevel sets the default log level threshold
	SetLogLevel(logLvl LogLevel)

	// SetOutput changes the output destination of the logger
	SetOutput(output io.Writer)

	// setFormatLogger use for to implemente the mock
	setFormatLogger(fl formatLogger)

	// SetRouter changes the router for this logger instance.  Once set, logs produced by this
	// logger will not be touched by the global router.  Mostly used for testing and benchmarking.
	SetRouter(router router.Router)

	//
	// Logging
	//

	// Counter takes a string and logs with LogLevel = Info
	Counter(title string)

	// CounterD takes a string, value, and data map. It logs with LogLevel = Info
	CounterD(title string, value int, data map[string]interface{})

	// Critical takes a string and logs with LogLevel = Critical
	Critical(title string)

	// CriticalD takes a string and data map. It logs with LogLevel = Critical
	CriticalD(title string, data map[string]interface{})

	// Debug takes a string and logs with LogLevel = Debug
	Debug(title string)

	// DebugD takes a string and data map. It logs with LogLevel = Debug
	DebugD(title string, data map[string]interface{})

	// Error takes a string and logs with LogLevel = Error
	Error(title string)

	// ErrorD takes a string and data map. It logs with LogLevel = Error
	ErrorD(title string, data map[string]interface{})

	// GaugeFloat takes a string and float value. It logs with LogLevel = Info
	GaugeFloat(title string, value float64)

	// GaugeFloatD takes a string, a float value, and data map. It logs with LogLevel = Info
	GaugeFloatD(title string, value float64, data map[string]interface{})

	// GaugeInt takes a string and integer value. It logs with LogLevel = Info
	GaugeInt(title string, value int)

	// GaugeIntD takes a string, an integer value, and data map. It logs with LogLevel = Info
	GaugeIntD(title string, value int, data map[string]interface{})

	// Info takes a string and logs with LogLevel = Info
	Info(title string)

	// InfoD takes a string and data map. It logs with LogLevel = Info
	InfoD(title string, data map[string]interface{})

	// Warn takes a string and logs with LogLevel = Warning
	Warn(title string)

	// WarnD takes a string and data map. It logs with LogLevel = Warning
	WarnD(title string, data map[string]interface{})
}

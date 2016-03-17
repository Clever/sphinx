# logger
--
    import "gopkg.in/Clever/kayvee-go.v3/logger"


## Usage

#### type Formatter

```go
type Formatter func(data map[string]interface{}) string
```

Formatter is a function type that takes a map and returns a formatted string
with the contents of the map

#### type LogLevel

```go
type LogLevel int
```

LogLevel is an enum is used to denote level of logging

```go
const (
	Debug LogLevel = iota
	Info
	Warning
	Error
	Critical
)
```
Constants used to define different LogLevels supported

#### func (LogLevel) String

```go
func (l LogLevel) String() string
```

#### type Logger

```go
type Logger struct {
}
```

Logger provides customization of log messages. We can change globals, default
log level, formatting, and output destination.

#### func  New

```go
func New(source string) *Logger
```
New creates a *logger.Logger. Default values are Debug LogLevel, kayvee
Formatter, and std.err output.

#### func (*Logger) Counter

```go
func (l *Logger) Counter(title string)
```
Counter takes a string and logs with LogLevel = Info, type = counter, and value
= 1

#### func (*Logger) CounterD

```go
func (l *Logger) CounterD(title string, value int, data map[string]interface{})
```
CounterD takes a string, value, and data map. It logs with LogLevel = Info, type
= counter, and value = value

#### func (*Logger) Critical

```go
func (l *Logger) Critical(title string)
```
Critical takes a string and logs with LogLevel = Critical

#### func (*Logger) CriticalD

```go
func (l *Logger) CriticalD(title string, data map[string]interface{})
```
CriticalD takes a string and data map. It logs with LogLevel = Critical

#### func (*Logger) Debug

```go
func (l *Logger) Debug(title string)
```
Debug takes a string and logs with LogLevel = Debug

#### func (*Logger) DebugD

```go
func (l *Logger) DebugD(title string, data map[string]interface{})
```
DebugD takes a string and data map. It logs with LogLevel = Debug

#### func (*Logger) Error

```go
func (l *Logger) Error(title string)
```
Error takes a string and logs with LogLevel = Error

#### func (*Logger) ErrorD

```go
func (l *Logger) ErrorD(title string, data map[string]interface{})
```
ErrorD takes a string and data map. It logs with LogLevel = Error

#### func (*Logger) GaugeFloat

```go
func (l *Logger) GaugeFloat(title string, value float64)
```
GaugeFloat takes a string and float value. It logs with LogLevel = Info, type =
gauge, and value = value

#### func (*Logger) GaugeFloatD

```go
func (l *Logger) GaugeFloatD(title string, value float64, data map[string]interface{})
```
GaugeFloatD takes a string, a float value, and data map. It logs with LogLevel =
Info, type = gauge, and value = value

#### func (*Logger) GaugeInt

```go
func (l *Logger) GaugeInt(title string, value int)
```
GaugeInt takes a string and integer value. It logs with LogLevel = Info, type =
gauge, and value = value

#### func (*Logger) GaugeIntD

```go
func (l *Logger) GaugeIntD(title string, value int, data map[string]interface{})
```
GaugeIntD takes a string, an integer value, and data map. It logs with LogLevel
= Info, type = gauge, and value = value

#### func (*Logger) Info

```go
func (l *Logger) Info(title string)
```
Info takes a string and logs with LogLevel = Info

#### func (*Logger) InfoD

```go
func (l *Logger) InfoD(title string, data map[string]interface{})
```
InfoD takes a string and data map. It logs with LogLevel = Info

#### func (*Logger) SetConfig

```go
func (l *Logger) SetConfig(source string, logLvl LogLevel, formatter Formatter, output io.Writer)
```
SetConfig allows configuration changes in one go

#### func (*Logger) SetFormatter

```go
func (l *Logger) SetFormatter(formatter Formatter)
```
SetFormatter sets the formatter function to use

#### func (*Logger) SetLogLevel

```go
func (l *Logger) SetLogLevel(logLvl LogLevel)
```
SetLogLevel sets the default log level threshold

#### func (*Logger) SetOutput

```go
func (l *Logger) SetOutput(output io.Writer)
```
SetOutput changes the output destination of the logger

#### func (*Logger) Warn

```go
func (l *Logger) Warn(title string)
```
Warn takes a string and logs with LogLevel = Warning

#### func (*Logger) WarnD

```go
func (l *Logger) WarnD(title string, data map[string]interface{})
```
WarnD takes a string and data map. It logs with LogLevel = Warning

#### type M

```go
type M map[string]interface{}
```

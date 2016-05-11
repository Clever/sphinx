# kayvee
--
    import "gopkg.in/Clever/kayvee-go.v3"

Package kayvee provides methods to output human and machine parseable strings,
with a "json" format.

## [Logger API Documentation](./logger)

* (gopkg.in/Clever/kayvee-go.v3/logger)[https://godoc.org/gopkg.in/Clever/kayvee-go.v3/logger]
* (gopkg.in/Clever/kayvee-go.v3/middleware)[https://godoc.org/gopkg.in/Clever/kayvee-go.v3/middleware]

## Example

```go
    package main

    import(
        "fmt"
        "time"

        "gopkg.in/Clever/kayvee-go.v3/logger"
    )

    func main() {
        myLogger := logger.New("myApp")

        // Simple debugging
        myLogger.Debug("Service has started")

        // Make a query and log its length
        query_start := time.Now()
        myLogger.GaugeFloat("QueryTime", time.Since(query_start).Seconds())

        // Output structured data
        myLogger.InfoD("DataResults", map[string]interface{}{"key": "value"})

        // You can use the M alias for your key value pairs
        myLogger.InfoD("DataResults", logger.M{"shorter": "line"})
    }
```


## Testing

Run `make test` to execute the tests

## Change log

- v3.0 - Removed sentry-go dependency
- v2.4 - Add kayvee-go/validator for asserting that raw log lines are in a valid kayvee format.
- v2.3 - Expose logger.M.
- v2.2 - Remove godeps.
- v2.1 - Add kayvee-go/logger with log level, counters, and gauge support
- v0.1 - Initial release.

## Backward Compatibility

The kayvee 1.x interface still exist but is considered deprecated. You can find documentation on using it in the [compatibility guide](./compatibility.md)


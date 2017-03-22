# kayvee
--
    import "gopkg.in/Clever/kayvee-go.v6"

Package kayvee provides methods to output human and machine parseable strings,
with a "json" format.

## [Logger API Documentation](./logger)

* [gopkg.in/Clever/kayvee-go.v6/logger](https://godoc.org/gopkg.in/Clever/kayvee-go.v6/logger)
* [gopkg.in/Clever/kayvee-go.v6/middleware](https://godoc.org/gopkg.in/Clever/kayvee-go.v6/middleware)

## Examples

```go
// main.go
package main

import (
    l "log"
    "path"
    "time"

    "github.com/kardianos/osext"
    "gopkg.in/Clever/kayvee-go.v6/logger"
)

var log = logger.New("myApp")

func init() {
    // Use osext library to consistently find kvconfig.yml file
    dir, err := osext.ExecutableFolder()
    if err != nil {
        l.Fatal(err)
    }
    err = logger.SetGlobalRouting(path.Join(dir, "kvconfig.yml"))
    if err != nil {
        l.Fatal(err)
    }
}

func main() {
    // Simple debugging
    log.Debug("Service has started")

    // Make a query and log its length
    query_start := time.Now()
    log.GaugeFloat("QueryTime", time.Since(query_start).Seconds())

    // Output structured data
    log.InfoD("DataResults", logger.M{"key": "value"})  // Sends slack message (see Log Routing)

    // You can use the M alias for your key value pairs
    log.InfoD("DataResults", logger.M{"shorter": "line"}) // will NOT send slack message
}
```

## Log Routing

Log routing is a mechanism for defining where log lines should go once they've entered Clever's logging pipeline.   Routes are defined in a yaml file called kvconfig.yml.  Here's an example of a log routing rule that sends a slack message:

```yaml
# kvconfig.yml
routes:
  key-val: # Rule name
    matchers:
      title: [ "DataResults", "QueryResults" ]
      key: [ "value" ]
    output: # Routes log line to #data-dinesty slack channel
      type: "notifications"
      channel: "#data-dinesty"
      icon: ":bird:"
      message: "The data is in: %{key}"
      user: "The Data Duck"
```

For more information see https://clever.atlassian.net/wiki/display/ENG/Log+Routing

## Testing

Run `make test` to execute the tests

### Testing Log Routing

A mock logger is provided to make it easier to test log routing rules.  Here's an exampe:

```go
// Units for main.go which is defined in the examples section of this README
package main

import (
    l "log"
    "testing"

    "github.com/stretchr/testify/assert"
    "gopkg.in/Clever/kayvee-go.v6/logger"
)

func init() {
    err := logger.SetGlobalRouting("./kvconfig.yml")
    if err != nil {
        l.Fatal(err)
    }
}

func TestDataResultsRouting(t *testing.T) {
    assert := assert.New(t)

    mocklog := logger.NewMockCountLogger("myApp")

    // Overrides package level logger
    log = mocklog

    main() // Call function to generate log lines

    counts := mocklog.RuleCounts()

    assert.Contains(counts, "key-val")
    assert.Equal(counts["key-val"], 1)
}
```


## Change log

- v6.0 - Introduced log-routing
- v5.0 - Middleware logger now creates a new logger on each request.
  - Breaking change to `middleware.New` constructor.
- v4.0
  - Added methods to read and write the `Logger` object from a a `context.Context` object.
  - Middleware now injects the logger into the request context.
  - Updated to require Go 1.7.
- v4.0 - Removed sentry-go dependency
- v2.4 - Add kayvee-go/validator for asserting that raw log lines are in a valid kayvee format.
- v2.3 - Expose logger.M.
- v2.2 - Remove godeps.
- v2.1 - Add kayvee-go/logger with log level, counters, and gauge support
- v0.1 - Initial release.

## Backward Compatibility

The kayvee 1.x interface still exist but is considered deprecated. You can find documentation on using it in the [compatibility guide](./compatibility.md)

## Publishing

To release a new version run `make bump-major`, `make bump-minor`, or `make
bump-patch` as appropriate on master (after merging your PR). Then, run `git
push --tags`.

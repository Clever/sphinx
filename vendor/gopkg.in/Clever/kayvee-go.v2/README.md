# kayvee
--
    import "gopkg.in/clever/kayvee-go.v2"

Package kayvee provides methods to output human and machine parseable strings,
with a "key=val" format.

## Example

Here's an example program that outputs a kayvee formatted string:

    package main

    import(
      "fmt"
      "gopkg.in/Clever/kayvee-go.v2"
    )

    func main() {
      fmt.Println(kayvee.Format(map[string]interface{}{"hello": "world"}))
    }

## Testing


Run `make test` to execute the tests

## Change log

v0.0.1 - Initial release.

## Usage

#### func  Format

```go
func Format(data map[string]interface{}) string
```
Format converts a map to a string of space-delimited key=val pairs

#### func  FormatLog

```go
func FormatLog(source string, level string, title string, data map[string]interface{}) string
```
FormatLog is similar to Format, but takes additional reserved params to promote
logging best-practices

# common
--
    import "github.com/Clever/sphinx/common"


## Usage

#### func  ConstructMockRequestWithHeaders

```go
func ConstructMockRequestWithHeaders(headers map[string][]string) *http.Request
```
ConstructMockRequestWithHeaders constructs an http.Request with the given
headers

#### func  Hash

```go
func Hash(str, salt string) string
```
Hash hashes a string based on the given salt

#### func  InSlice

```go
func InSlice(a string, list []string) bool
```
InSlice tests whether or not a string exists in a slice of strings

#### func  ReMarshal

```go
func ReMarshal(config interface{}, target interface{}) error
```
ReMarshal parses interface{} into concrete types

#### func  SortedKeys

```go
func SortedKeys(obj map[string]interface{}) []string
```
SortedKeys returns a sorted slice of map keys

#### type Request

```go
type Request map[string]interface{}
```

Request contains any info necessary to ratelimit a request

#### func  HTTPToSphinxRequest

```go
func HTTPToSphinxRequest(r *http.Request) Request
```
HTTPToSphinxRequest converts an http.Request to a Request

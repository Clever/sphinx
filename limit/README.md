# limit
--
    import "github.com/Clever/sphinx/limit"


## Usage

#### type Limit

```go
type Limit interface {
	Name() string
	Match(common.Request) bool
	Add(common.Request) (leakybucket.BucketState, error)
}
```

Limit has methods for matching and adding to a limit

#### func  New

```go
func New(name string, config config.Limit, storage leakybucket.Storage) (Limit, error)
```
New creates a new Limit

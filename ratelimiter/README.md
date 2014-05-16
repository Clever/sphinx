# ratelimiter
--
    import "github.com/Clever/sphinx/ratelimiter"


## Usage

#### type RateLimiter

```go
type RateLimiter interface {
	Add(request common.Request) ([]Status, error)
}
```

RateLimiter rate limits requests based on given configuration and limits.

#### func  New

```go
func New(config config.Config) (RateLimiter, error)
```
New returns a new RateLimiter based on the given configuration.

#### type Status

```go
type Status struct {
	Capacity  uint
	Reset     time.Time
	Remaining uint
	Name      string
}
```

Status contains the status of a limit.

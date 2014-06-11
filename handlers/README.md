# handlers
--
    import "github.com/Clever/sphinx/handlers"


## Usage

#### type SphinxHandler

```go
type SphinxHandler interface {
	http.Handler
	SetRateLimiter(rateLimiter ratelimiter.RateLimiter)
}
```


#### func  NewHTTPLimiter

```go
func NewHTTPLimiter(rateLimiter ratelimiter.RateLimiter, proxy http.Handler) SphinxHandler
```
NewHTTPLimiter returns an http.Handler that rate limits and proxies requests.

#### func  NewHTTPLogger

```go
func NewHTTPLogger(rateLimiter ratelimiter.RateLimiter, proxy http.Handler) SphinxHandler
```
NewHTTPLogger returns an http.Handler that logs the results of rate limiting
requests, but actually proxies everything.

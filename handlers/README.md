# handlers
--
    import "github.com/Clever/sphinx/handlers"


## Usage

#### func  NewHTTPLimiter

```go
func NewHTTPLimiter(rateLimiter ratelimiter.RateLimiter, proxy http.Handler) http.Handler
```
NewHTTPLimiter returns an http.Handler that rate limits and proxies requests.

#### func  NewHTTPLogger

```go
func NewHTTPLogger(rateLimiter ratelimiter.RateLimiter, proxy http.Handler) http.Handler
```
NewHTTPLogger returns an http.Handler that logs the results of rate limiting
requests, but actually proxies everything.

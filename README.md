## Sphinx: HTTP Rate Limiting

_Sphinx_ is a _rate limiting_ HTTP proxy, implemented in Go, using
[leaky buckets](https://github.com/Clever/leakybucket).

*The name for this project (_"Sphinx"_) comes from the ancient Greek word sphingien, which means "to squeeze" or "to strangle." 
The Sphinx would stand by the road and stop travelers to ask them a riddle.
If they could not answer, she would strangle them. She was often thought of as a guardian and flanked the entrances to temples.*

![Sphinx](logo.png)

## Why?

Rate limiting an API is often required to ensure that clients do not abuse the available resources and that the API is reliably available when multiple clients are requesting data concurrently.
Buckets can be created based on various parameters of an incoming request (eg. Authorization, IP address) to configure how requests are grouped for limiting.

Rate limiting functionality is already available in some proxies (eg. Nginx, HAProxy).
However, they often use in-memory stores that make rate-limiting when running multiple proxies (e.g. for load balancing) unpredictable.
Configuration for these limits also gets complex since it includes many actions such as routing, request/response re-writing, and rate-limiting.

## Sphinx is not...

* Sphinx is not focused on preventing Denial of Service (DoS) attacks or requests from malicious clients.
The goal is to expose rate limiting information to clients and enforce balanced use by API clients.

* Sphinx is not a request forwarding service.
Sphinx only allows for very simplistic forwarding to a single host per instance of the rate limiter.
Any advanced routing or request handling should be handled by a _real_ proxy (eg. Nginx, HAProxy).

* Sphinx is not an HTTPS terminator.
This keeps the burden of configuring _SSL certificates_ and security outside of Sphinx.
Ideally, there is real load balancing and HTTPS termination before a request hits Sphinx.

## Rate limit headers and errors

_Sphinx_ will update HTTP response headers for requests that *match limits* to include details about the rate limit status.
Headers are _[canonicalized](http://golang.org/pkg/net/http/#CanonicalHeaderKey)_, but clients should assume [header names are case insensitive](http://www.w3.org/Protocols/rfc2616/rfc2616-sec4.html#sec4.2).

 - _X-RateLimit-Reset_: Unix timestamp when the rate limit counter will be reset.
 - _X-RateLimit-Limit_: The total number of requests allowed in a time period.
 - _X-RateLimit-Remaining_: Number of requests that can be made until the reset time.
 - _X-RateLimit-Bucket_: Name of the rate-limit bucket this request belongs to in the configuration.

_Limit names_ can be configured via a configuration file.

Request:

    HOST example.com
    GET /resource/123
    AUTHORIZATION Basic ABCD

Response headers:

    Status: 200 OK
    X-RateLimit-Limit: 200
    X-RateLimit-Remaining: 199
    X-RateLimit-Reset: 1394506274
    X-RateLimit-Bucket: authorized-users

In case the client hits a rate limit, an empty response with a `429 Too Many
Requests` status code will be returned.

Request:

    HOST example.com
    GET /resource/123
    AUTHORIZATION Basic ABC

Response headers:

    Status: 429 Too Many Requests
    X-RateLimit-Limit: 200
    X-RateLimit-Remaining: 0
    X-RateLimit-Reset: 1394506274
    X-RateLimit-Bucket: authorized-users

## Documentation

[![LeakyBucket](https://godoc.org/github.com/Clever/leakybucket?status.png)](https://godoc.org/github.com/Clever/leakybucket).
[![Sphinx](https://godoc.org/github.com/Clever/sphinx?status.png)](https://godoc.org/github.com/Clever/sphinx).

## Tests

_Sphinx_ is built and tested against Go 1.2.
Ensure this is the version of Go you're running with `go version`.
Make sure your GOPATH is set, e.g. `export GOPATH=~/go`.
Clone the repository to a location outside your GOPATH, and symlink it to `$GOPATH/src/github.com/Clever/sphinx`.
If you have [gvm](https://github.com/moovweb/gvm) installed, you can make this symlink by running the following from the root of the repository: `gvm linkthis github.com/Clever/sphinx`.

If you have done all of the above, then you should be able to run

```
make test
```

If you'd like to see a code coverage report, install the cover tool 
(`go get code.google.com/p/go.tools/cmd/cover`), make sure `$GOPATH/bin` 
is in your PATH, and run:

```
COVERAGE=1 make
```

## Credits

* [Sphinx](http://thenounproject.com/term/sphinx/20572/) logo by EricP from The Noun Project
* [Drone](https://github.com/drone/drone) inspiration for building a deb

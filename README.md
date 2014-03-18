## Sphinx: HTTP Rate Limiting

_Sphinx_ is a _rate limiting_ HTTP proxy implemented in Go built using 
[leaky buckets](https://github.com/Clever/leakybucket).

The name for this project _"Sphinx"_ comes from the ancient Greek word sphingien, which means "to squeeze" or "to strangle." 
Sphinx would stand by the road and stop passers by. She would then ask them a riddle. If they could not answer, 
she would strangle them. Sphinx was thought of as a guardian often flanking the entrances to temples.

![Ancient Greek sphinx from Delphi](https://upload.wikimedia.org/wikipedia/commons/thumb/2/20/028MAD_Sphinx.jpg/173px-028MAD_Sphinx.jpg)

## Why

Rate limiting API's is often required to ensure that clients do not abuse the available resources
and that the API is reliably available when multiple clients are requesting data concurrently. Buckets
can be created based on various parameters of an incoming request (eg. Authorization, ip address etc) to
configure how requests are grouped for limiting.

Rate limiting functionalities are available in some proxies (eg. Nginx, HAProxy) but they often use in-memory
stores that make predictable rate-limiting really hard when running multiple proxies for load balancing 
or reliability.  The configuration for setting up these limits also gets complex since it includes 
many actions including routing, request/response re-writing, rate-limiting.

## What Sphinx is not?

* This service is NOT focused on preventing Denial of Service (DoS)
attacks or requests from malicious clients. The goal is to allow the API to
transparently exposing rate limiting information to clients and enforce
balanced use by partners.

* Sphinx only allows for very simplistic forwarding. This would
primarily be forwarding to a single host per instance of the rate limiter. Any
advanced routing or request handling should be handled by a _real_ proxy (eg.
Nginx, HAProxy).


## Rate Limit Headers and Errors

_Sphinx_ will update the headers on HTTP responses to include details
about the rate limit status along with the response from the backend.

 - _X-Rate-Limit-Reset_: Unix timestamp when the rate limit counter will be reset.
 - _X-Rate-Limit-Limit_: The total number of requests allowed in a time period. 
 - _X-Rate-Limit-Remaining_: Number of requests that can be made until the reset time.
 - _X-Rate-Limit-Bucket_: Name of the rate-limit bucket this request belongs to in the configuration.

_bucket names_ and _response body_ can be set in the _Sphinx Configuration_.

Request:

    HOST example.com
    GET /resource/123
    AUTHORIZATION Basic ABCD

Response Headers:

    Status: 200 OK
    X-Rate-Limit-Limit: 200
    X-Rate-Limit-Remaining: 199
    X-Rate-Limit-Reset: 1394506274
    X-Rate-Limit-Bucket: authorized-users

In case your application hits a rate limit, a HTTP Status Code `429 Too Many
Requests` with an error message will be returned.

Request:

    HOST example.com
    GET /resource/123
    AUTHORIZATION Basic ABC

Response Headers:

    Status: 429 Too Many Requests
    X-Rate-Limit-Limit: 200
    X-Rate-Limit-Remaining: 0
    X-Rate-Limit-Reset: 1394506274
    X-Rate-Limit-Bucket: authorized-users

Response body:

    {
        "error": "API rate limit reached."
    }

## Documentation

[![GoDoc](https://godoc.org/github.com/Clever/leakybucket?status.png)](https://godoc.org/github.com/Clever/leakybucket).

## Tests

_Sphinx_ is built and tested against Go 1.2.
Ensure this is the version of Go you're running with `go version`.
Make sure your GOPATH is set, e.g. `export GOPATH=~/go`.
Clone the repository to a location outside your GOPATH, and symlink it to 
`$GOPATH/src/github.com/Clever/sphinx`.
If you have [gvm](https://github.com/moovweb/gvm) installed, you can 
make this symlink by running the following from the root of where you 
have cloned the repository: `gvm linkthis github.com/Clever/sphinx`.

If you have done all of the above, then you should be able to run

```
make
```

If you'd like to see a code coverage report, install the cover tool 
(`go get code.google.com/p/go.tools/cmd/cover`), make sure `$GOPATH/bin` 
is in your PATH, and run:

```
COVERAGE=1 make
```

If you'd like to see lint your code, install golint (`go get github.com/golang/lint/golint`) and run:

```
LINT=1 make
```

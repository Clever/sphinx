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

## [Configuring Sphinx](./example.yaml)

Rate limiting in Sphinx is managed by setting up `limits` in a `yaml` configuration file. 
Details about the configuration format can be found in the [annotated example](./example.yaml).

It is important to understand the concept of `buckets` and `limits` to effectively configure a rate limiter.

_Limit_: A limit defines a rate limiting policy that Sphinx enforces by counting requests in named _buckets_.
_Bucket_: A bucket is simply a named value. Each request that matches a limit increments the value of one bucket.

Below is an example of a limit and three requests that increment two bucket values.


### Test Limit

  match if request path begins with `/limited`
  bucket names are defined as `name-{ip-address}`
  Allow TWO requests per minute

Setting this limit using the config would look like: 

```
proxy:
  handler: http             # can be {http,httplogger}
  host: http://httpbin.org  # URI for the http(s) backend we are proxying to
  listen: :6634             # bind to host:port. default: height of the Great Sphinx of Giza

storage:
  type: memory    # can be {redis,memory}

limits:
  test-limit:
    interval: 60  # in seconds
    max: 2        # number of requests allowed in interval
    keys:
      ip: ""      # ip keys require no configuration
    matches:
      paths:
        match_any:
          - "/limited*"
```

### Request One

  path: /limited/resource/1
  Headers:
    Host: example.com
    Authorization: Basic User:Password
    IP: 10.0.0.1

*State*:
    `test-limit-10.0.0.1`: 1

### Request Two

  path: /limited/resource/2
  Headers:
    Host: example.com
    Authorization: Basic Admin:Secure
    IP: 10.0.0.2

*State*
    `test-limit-10.0.0.1`: 1
    `test-limit-10.0.0.2`: 1


### Request Three

  path: /limited/resource/3
  Headers:
    Host: example.com
    Authorization: Basic Admin:Secure
    IP: 10.0.0.1

*State*
    `test-limit-10.0.0.1`: 2
    `test-limit-10.0.0.2`: 1


The following snippet explains how to define limits in Sphinx:

```yaml
limit-name:
  interval: 15
  max: 200
  keys:
    headers:
      names:
        - "Authorization"
  matches:
    paths:
      match_any:
        - "/special/resources/.*"
```

_limit\_name_: Used to identify and added to the _X-RateLimit-Bucket_ header.

_interval_: A limit may create many `buckets`. This key provides the `expire time in secs` for all
            buckets created for this limit.

_max_: Maximum number of requests that will be allowed for a `bucket` in one `interval`.

_keys_: This section defines the dynamic bucket name generated for each request. Currently supported matchers include `headers` and `ip`. 
All keys defined are concatenated to create the full bucket name.

  _headers_: Use concatenated header values from requests in the `bucket` name.
```yaml
headers:
  encrypt: "SALT_TO_ENCRYPT_VALUE"  # optional
  names:
    - HEADER_NAME_1
    - HEADER_NAME_2
```

  _ip_: Use the incoming `IP Address` from the incoming request in the bucket name.

_matches_: This section defines which requests this limit should be applied to. The request `MUST` match _all_ of the matchers defined in this
block. Currently supported matchers are `headers` and `paths`.

  _headers_: This matcher currently supports the `match_any` key which returns true if _any_ of the list items evaluate to true. eg:
```yaml
headers:
  match_any:
    - name: "HEADER_NAME"
      match: "REGEX_FOR_MATCHING_HEADER_VALUE"
    - name: "OTHER_HEADER_NAME"  # no match key means just check for existence
```

  _paths_: This matcher also supports the `match_any` key.
```yaml
paths:
  match_any:
    - "/limited/resource/*"
    - "/objects/limited/.*"
```


## Documentation

  * LeakyBucket:  [![LeakyBucket documentation](https://godoc.org/github.com/Clever/leakybucket?status.png)](https://godoc.org/github.com/Clever/leakybucket)
  * Sphinx: [![Sphinx documentation](https://godoc.org/github.com/Clever/sphinx?status.png)](https://godoc.org/github.com/Clever/sphinx)


## Tests

_Sphinx_ is built and tested against Go 1.5.
Ensure this is the version of Go you're running with `go version`.
Make sure your GOPATH is set, e.g. `export GOPATH=~/go`.

```bash
mkdir -p $GOPATH/src/github.com/Clever
cd $GOPATH/src/github.com/Clever
git clone git@github.com:Clever/sphinx.git
```

Now you can run our test and linting suites via Make:
```
cd sphinx
make test
```


## Credits

* [Sphinx](http://thenounproject.com/term/sphinx/20572/) logo by EricP from The Noun Project
* [Drone](https://github.com/drone/drone) inspiration for building a deb


## Vendoring

Please view the [dev-handbook for instructions](https://github.com/Clever/dev-handbook/blob/master/golang/godep.md).

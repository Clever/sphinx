# config
--
    import "github.com/Clever/sphinx/config"


## Usage

#### func  ValidateConfig

```go
func ValidateConfig(config Config) error
```
ValidateConfig validates that a Config has all the required fields TODO (z):
These should all be private, but right now tests depend on parsing bytes into
yaml

#### type Config

```go
type Config struct {
	Proxy       Proxy
	HealthCheck HealthCheck `yaml:"health-check"`
	Limits      map[string]Limit
	Storage     map[string]string
}
```

Config holds the yaml data for the config file

#### func  LoadAndValidateYaml

```go
func LoadAndValidateYaml(data []byte) (Config, error)
```
LoadAndValidateYaml turns a sequence of bytes into a Config and validates that
all the necessary fields are set TODO (z): These should all be private, but
right now tests depend on parsing bytes into yaml

#### func  LoadYaml

```go
func LoadYaml(data []byte) (Config, error)
```
LoadYaml loads byte data for a yaml file into a Config TODO (z): These should
all be private, but right now tests depend on parsing bytes into yaml

#### func  New

```go
func New(path string) (Config, error)
```
New takes in a path to a configuration yaml and returns a Configuration.

#### type HealthCheck

```go
type HealthCheck struct {
	Port     string
	Endpoint string
}
```

HealthCheck holds the yaml data for how to run the health check service.

#### type Limit

```go
type Limit struct {
	Interval uint
	Max      uint
	Keys     map[string]interface{}
	Matches  map[string]interface{}
	Excludes map[string]interface{}
}
```

Limit holds the yaml data for one of the limits in the config file

#### type Proxy

```go
type Proxy struct {
	Handler string
	Host    string
	Listen  string
}
```

Proxy holds the yaml data for the proxy option in the config file

# daemon
--
    import "github.com/Clever/sphinx/daemon"


## Usage

#### type Daemon

```go
type Daemon interface {
	Start()
}
```


#### func  New

```go
func New(config config.Config) (Daemon, error)
```
NewDaemon takes in config.Configuration and creates a sphinx listener

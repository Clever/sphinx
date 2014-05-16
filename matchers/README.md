# matchers
--
    import "github.com/Clever/sphinx/matchers"


## Usage

#### type Matcher

```go
type Matcher interface {
	Match(common.Request) bool
}
```

A Matcher determines if a common.Request matches its requirements.

#### type MatcherFactory

```go
type MatcherFactory interface {
	Type() string
	Create(config interface{}) (Matcher, error)
}
```

A MatcherFactory creates a Matcher based on a config.

#### func  MatcherFactoryFinder

```go
func MatcherFactoryFinder(name string) MatcherFactory
```
MatcherFactoryFinder finds a MatcherFactory by name.

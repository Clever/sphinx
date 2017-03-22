package router

//go:generate ./generate_schema.sh

// Router is an an interface for an object that can route log lines.
type Router interface {
	Route(map[string]interface{}) map[string]interface{}
}

// RuleRouter is an object that can route log lines according to `rules`.
type RuleRouter struct {
	rules []Rule
}

// RuleMatchers describes which log lines a router rule applies to.
type RuleMatchers map[string][]string

// RuleOutput describes what to do if a log line matches a rule.
type RuleOutput map[string]interface{}

// Rule is a log routing rule
type Rule struct {
	Name     string       `json:"-"`
	Matchers RuleMatchers `json:"matchers"`
	Output   RuleOutput   `json:"output"`
}

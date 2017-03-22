package router

import (
	"fmt"
	"io/ioutil"
	"os"

	kv "gopkg.in/Clever/kayvee-go.v6"
)

var teamName string

func init() {
	teamName = os.Getenv("_TEAM_OWNER")
	if teamName == "" {
		teamName = "UNSET"
	}
}

func setDefaults(output map[string]interface{}) map[string]interface{} {
	otype, ok := output["type"].(string)
	if !ok {
		return output
	}

	switch otype {
	case "metrics":
		fallthrough
	case "alerts":
		if _, ok := output["value_field"]; !ok {
			output["value_field"] = "value"
		}
	}

	return output
}

// Route returns routing metadata for the log line `msg`. The outputs (with
// variable substitutions performed) for each rule matched are placed under the
// "routes" key.
func (r *RuleRouter) Route(msg map[string]interface{}) map[string]interface{} {
	outputs := []map[string]interface{}{}
	for _, rule := range r.rules {
		if rule.Matches(msg) {
			outputs = append(outputs, rule.OutputFor(msg))
		}
	}
	return map[string]interface{}{
		"team":        teamName,
		"kv_version":  kv.Version,
		"kv_language": "go",
		"routes":      outputs,
	}
}

// NewFromConfig constructs a Router using the configuration specified as yaml
// in `filename`. The routing rules should be placed under the "routes" key on
// the root-level map in the file. Validation is performed as described in
// parse.go.
func NewFromConfig(filename string) (Router, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	router, err := newFromConfigBytes(fileBytes)
	if err != nil {
		return nil, fmt.Errorf(
			"Error initializing kayvee log router from file '%s':\n%s",
			filename, err.Error(),
		)
	}
	return router, nil
}

func newFromConfigBytes(fileBytes []byte) (Router, error) {
	routes, err := parse(fileBytes)
	if err != nil {
		return &RuleRouter{}, err
	}

	return NewFromRoutes(routes)
}

// NewFromRoutes constructs a RuleRouter using the provided map of route names
// to Rules.
func NewFromRoutes(routes map[string]Rule) (Router, error) {
	router := &RuleRouter{}
	for name, rule := range routes {
		output, err := substituteEnvVars(rule.Output)
		if err != nil {
			return router, err
		}
		output = setDefaults(output)

		rule.Name = name
		rule.Output = output
		router.rules = append(router.rules, rule)
	}

	return router, nil
}

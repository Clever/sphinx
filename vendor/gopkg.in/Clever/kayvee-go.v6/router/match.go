package router

import (
	"strings"
)

// Matches returns true if the `msg` matches the matchers specified in this
// routing rule.
func (r *Rule) Matches(msg map[string]interface{}) bool {
	for field, values := range r.Matchers {
		if !fieldMatches(field, values, msg) {
			return false
		}
	}
	return true
}

// OutputFor returns the output map for this routing rule with substitutions
// applied in accordance with the current environment and the contents of the
// message.
func (r *Rule) OutputFor(msg map[string]interface{}) map[string]interface{} {
	lookup := func(field string) (interface{}, bool) {
		return lookupField(field, msg)
	}
	subbed := substituteFields(r.Output, lookup)
	subbed["rule"] = r.Name
	return subbed
}

// lookupField does an extended lookup on `obj`, interpreting dots in field as
// corresponding to subobjects. It returns the value and true if the lookup
// succeeded or `"", false` if the key is missing or corresponds to a
// non-string value.
func lookupField(field string, obj map[string]interface{}) (interface{}, bool) {
	if strings.Index(field, ".") == -1 {
		val, ok := obj[field]
		return val, ok
	}
	return lookupFieldPath(strings.Split(field, "."), obj)
}

// lookupFieldPath does an extended lookup on `obj`, with each entry in `fieldPath`
// corresponding to subobjects. It returns the value and true if the lookup
// succeeded or `"", false` if a key was missing along the path or if the final
// key corresponds to a non-string value.
func lookupFieldPath(fieldPath []string, obj map[string]interface{}) (interface{}, bool) {
	part := fieldPath[0]
	if len(fieldPath) == 1 {
		val, ok := obj[part]
		return val, ok
	}
	if subObj, ok := obj[part].(map[string]interface{}); ok {
		return lookupFieldPath(fieldPath[1:], subObj)
	}
	return "", false
}

// fieldMatches returns true if the value of the key `field` in the map `obj`
// is one of `values`. Dots in `field` are interpreted as denoting subobjects
// -- i.e. the field name "x.y.z" says to check obj["x"]["y"]["z"].
func fieldMatches(field string, valueMatchers []string, obj map[string]interface{}) bool {
	val, ok := lookupField(field, obj)
	if !ok {
		return false
	}

	strVal := ""
	switch v := val.(type) {
	case nil:
		return false
	case string:
		strVal = v
	case bool:
		if v {
			strVal = "true"
		} else {
			strVal = "false"
		}
	default: // Wildcard should match anything that isn't null or ""
		return valueMatchers[0] == "*"
	}

	if strVal == "" {
		return false
	}
	if valueMatchers[0] == "*" {
		return true
	}
	for _, match := range valueMatchers {
		if strVal == match {
			return true
		}
	}
	return false
}

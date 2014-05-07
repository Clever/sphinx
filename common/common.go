package common

// Request contains any info necessary to ratelimit a request
type Request map[string]interface{}

// InSlice tests whether or not a string exists in a slice of strings
func InSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

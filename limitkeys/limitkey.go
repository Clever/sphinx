package limitkeys

import (
	"fmt"

	"github.com/Clever/sphinx/common"
)

// A LimitKey returns a string key based on the request for creating bucketnames.
type LimitKey interface {
	Type() string
	Key(common.Request) (string, error)
}

// A EmptyKeyError signifies that the request does not contain enough information
// to create a key.
type EmptyKeyError struct {
	limitkey LimitKey
	message  string
}

func (eke EmptyKeyError) Error() string {
	return fmt.Sprintf("LimitKeyType: %s. %s", eke.limitkey.Type(), eke.message)
}

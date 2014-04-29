package limitkeys

import (
	"fmt"
	"github.com/Clever/sphinx/common"
)

type LimitKey interface {
	Key(common.Request) (string, error)
	Type() string
}

type EmptyKeyError struct {
	limitkey LimitKey
	message  string
}

func (eke EmptyKeyError) Error() string {
	return fmt.Sprintf("LimitKeyType: %s. %s", eke.limitkey.Type(), eke.message)
}

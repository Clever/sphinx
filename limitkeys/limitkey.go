package limitkeys

import "github.com/Clever/sphinx/common"

type LimitKey interface {
	GetKey(common.Request) string
}

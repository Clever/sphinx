package limitkeys

import (
	"github.com/Clever/sphinx/common"
)

const globalLimitKeyValue = "global:sigelton-key"

type globalLimitKey struct{}

func (ilk globalLimitKey) Type() string {
	return "global"
}

func (ilk globalLimitKey) Key(request common.Request) (string, error) {

	return globalLimitKeyValue, nil
}

// NewGlobalLimitKey creates a slice of globalLimitKeys that always returns the same key
func NewGlobalLimitKey(config interface{}) ([]LimitKey, error) {
	return []LimitKey{globalLimitKey{}}, nil
}

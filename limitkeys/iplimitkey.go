package limitkeys

import (
	"github.com/Clever/sphinx/common"
)

type ipLimitKey struct{}

func (ilk ipLimitKey) Type() string {
	return "ip"
}

func (ilk ipLimitKey) Key(request common.Request) (string, error) {

	if _, ok := request["remoteaddr"]; !ok {
		return "", EmptyKeyError{ilk, "No remoteaddr key in request"}
	}

	return "ip:" + request["remoteaddr"].(string), nil
}

// NewIPLimitKeys creates a sliced of ipLimitKeys that returns a key based on request remoteaddr
func NewIPLimitKeys(config interface{}) ([]LimitKey, error) {
	return []LimitKey{ipLimitKey{}}, nil
}

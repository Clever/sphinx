package limitkeys

import (
	"fmt"
	"github.com/Clever/sphinx/common"
)

type ipLimitKey struct {
}

func (ilk ipLimitKey) Type() string {
	return "ip"
}

func (ilk ipLimitKey) Key(request common.Request) (string, error) {

	if _, ok := request["remoteaddr"]; !ok {
		return "", EmptyKeyError{ilk, "No remoteaddr key in request"}
	}

	return fmt.Sprintf("ip:%s", request["remoteaddr"]), nil
}

// NewIPLimitKey creates a ipLimitKey that returns a key based on request remoteaddr
func NewIPLimitKey() ipLimitKey {
	return ipLimitKey{}
}

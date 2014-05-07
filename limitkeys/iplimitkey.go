package limitkeys

import (
	"github.com/Clever/sphinx/common"
)

type IPLimitKey struct {
}

func (ilk IPLimitKey) Type() string {
	return "ip"
}

func (ilk IPLimitKey) Key(request common.Request) (string, error) {

	if _, ok := request["remoteaddr"]; !ok {
		return "", EmptyKeyError{ilk, "No remoteaddr key in request"}
	}

	return "ip:" + request["remoteaddr"].(string), nil
}

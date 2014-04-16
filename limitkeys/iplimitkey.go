package limitkeys

import (
	"fmt"
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

	return fmt.Sprintf("%s:%s", "ip", request["remoteaddr"]), nil
}

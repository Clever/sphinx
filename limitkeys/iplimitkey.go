package limitkeys

import (
	"fmt"
	"github.com/Clever/sphinx/common"
)

type IPLimitKey struct {
}

func (ilk IPLimitKey) GetKey(request common.Request) string {

	if _, ok := request["remoteaddr"]; !ok {
		return ""
	}

	return fmt.Sprintf("%s:%s", "remoteaddr", request["remoteaddr"])
}

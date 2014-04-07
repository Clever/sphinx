package limitkeys

import (
	"fmt"
	"github.com/Clever/sphinx/common"
	"strings"
)

type HeaderLimitKey struct {
	name string
}

func (hlk HeaderLimitKey) GetKey(request common.Request) string {

	if _, ok := request["headers"]; !ok {
		return ""
	}

	headers := request["headers"].(map[string][]string)

	if _, ok := headers[hlk.name]; !ok {
		return ""
	}

	return fmt.Sprintf("%s-%s", hlk.name,
		strings.Join(headers[hlk.name], ":"))
}

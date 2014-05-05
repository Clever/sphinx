package limitkeys

import (
	"fmt"
	"github.com/Clever/sphinx/common"
	"net/http"
	"strings"
)

type HeaderLimitKey struct {
	Name string
}

func (hlk HeaderLimitKey) Type() string {
	return "header"
}

func (hlk HeaderLimitKey) Key(request common.Request) (string, error) {

	if _, ok := request["headers"]; !ok {
		return "", EmptyKeyError{hlk, "No headers in request"}
	}

	headers := request["headers"].(http.Header)

	if _, ok := headers[hlk.Name]; !ok {
		return "", EmptyKeyError{hlk,
			fmt.Sprintf("Header %s not found in request", hlk.Name)}
	}

	return fmt.Sprintf("%s:%s", hlk.Name,
		strings.Join(headers[hlk.Name], ";")), nil
}

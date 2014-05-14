package limitkeys

import (
	"fmt"
	"github.com/Clever/sphinx/common"
	"net/http"
	"sort"
	"strings"
)

type headerLimitKey struct {
	name string
	salt string
}

func (hlk headerLimitKey) Type() string {
	return "header"
}

func (hlk headerLimitKey) Key(request common.Request) (string, error) {

	if _, ok := request["headers"]; !ok {
		return "", EmptyKeyError{hlk, "No headers in request"}
	}

	headers := request["headers"].(http.Header)

	if _, ok := headers[hlk.name]; !ok {
		return "", EmptyKeyError{hlk,
			fmt.Sprintf("Header %s not found in request", hlk.name)}
	}

	return fmt.Sprintf("%s:%s", hlk.name,
		common.Hash(strings.Join(headers[hlk.name], ";"), hlk.salt)), nil
}

type headerConfig struct {
	Names   []string
	Encrypt string
}

// NewHeaderLimitKeys creates a slice of headerLimitKeys that keys on the named request header
func NewHeaderLimitKeys(config interface{}) ([]LimitKey, error) {
	conf := headerConfig{}
	err := common.ReMarshal(config, &conf)
	if err != nil {
		return nil, err
	}
	keys := []LimitKey{}
	sort.Strings(conf.Names) // order of config should not change bucketnames
	for _, header := range conf.Names {
		keys = append(keys, &headerLimitKey{name: header, salt: conf.Encrypt})
	}
	return keys, nil
}

package matcher

import (
	"github.com/Clever/sphinx"
)

type HeaderMatcherConfig struct {
	HeaderName string
	Matches    []string
}

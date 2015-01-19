package ratelimiter

import (
	"github.com/Clever/sphinx/common"
	"time"
)

// START OMIT
// A RateLimiter adds a request to a rate limiting bucket and returns the result.
type RateLimiter interface {
	Add(request common.Request) ([]Status, error)
}

// A Status contains the result of adding a request to our limiting buckets.
type Status struct {
	Capacity  uint
	Reset     time.Time
	Remaining uint
	Name      string
}

// END OMIT

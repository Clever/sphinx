package ratelimiter

import (
	"github.com/Clever/sphinx/common"
)

// START OMIT
import (
	"github.com/stretchr/testify/mock"
)

type MockRateLimiter struct {
	*mock.Mock
}

func (r *MockRateLimiter) Add(request common.Request) ([]Status, error) {
	args := r.Mock.Called(request)
	return args.Get(0).([]Status), args.Error(1)
}

// Verify at compile time that MockRateLimiter implements the RateLimiter interface
var _ RateLimiter = &MockRateLimiter{}

// END OMIT

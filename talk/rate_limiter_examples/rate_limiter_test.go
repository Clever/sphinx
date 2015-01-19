package ratelimiter

import (
	"errors"
	"github.com/Clever/sphinx/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

// START OMIT
func DoSomethingWithRateLimiter(limiter RateLimiter) error {
	_, err := limiter.Add(common.Request{"path": "/debug"})
	return err
}

var tests = []struct{ error }{
	{error: nil}, {error: errors.New("garbage")},
}

func TestDoingSomethingWithRateLimiter(t *testing.T) {
	for _, test := range tests {
		limiter := &MockRateLimiter{Mock: new(mock.Mock)}
		limiter.On("Add", common.Request{"path": "/debug"}).Return([]Status{}, test.error)

		assert.Equal(t, DoSomethingWithRateLimiter(limiter), test.error)

		limiter.AssertExpectations(t)
	}
}

// END OMIT

package daemon

import (
	"github.com/Clever/sphinx/config"
	"github.com/Clever/sphinx/ratelimiter"
	"github.com/stretchr/testify/mock"
	"net/http"
	"testing"
)

type MockHandler struct {
	mock.Mock
}

func (h MockHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	_ = h.Mock.Called(rw, r)
}

var anyRateLimiter = mock.AnythingOfTypeArgument("*ratelimiter.rateLimiter")

func (h MockHandler) SetRateLimiter(r ratelimiter.RateLimiter) {
	_ = h.Mock.Called(r)
}

func TestConfigReload(t *testing.T) {
	conf, _ := config.New("../example.yaml")
	d := daemon{}
	mHandler := new(MockHandler)
	mHandler.Mock.On("SetRateLimiter", anyRateLimiter).Return(nil).Once()
	d.handler = mHandler
	d.ReloadConfig(conf)
	if d.rateLimiter == nil {
		t.Fatal("Didn't assign rate limiter")
	}
}

func TestFailedReload(t *testing.T) {
	conf, _ := config.New("../example.yaml")
	daemon, _ := New(conf)
	conf2 := config.Config{}
	err := daemon.ReloadConfig(conf2)
	if err == nil {
		t.Fatal("Should have errored on empty configuration")
	}
}

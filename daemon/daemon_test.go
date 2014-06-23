package daemon

import (
	"github.com/Clever/sphinx/config"
	"testing"
)

func TestConfigReload(t *testing.T) {
	conf, err := config.New("../example.yaml")
	if err != nil {
		t.Fatal("Error loading config: " + err.Error())
	}
	d := daemon{}
	d.LoadConfig(conf)
	if d.handler == nil {
		t.Fatal("Didn't assign handler")
	}
}

func TestFailedReload(t *testing.T) {
	conf, err := config.New("../example.yaml")
	if err != nil {
		t.Fatal("Error loading config: " + err.Error())
	}
	daemon, err := New(conf)
	if err != nil {
		t.Fatal("Error creating new daemon: " + err.Error())
	}
	conf2 := config.Config{}
	err = daemon.LoadConfig(conf2)
	if err == nil {
		t.Fatal("Should have errored on empty configuration")
	}

	conf.Proxy.Listen = ":1000"
	err = daemon.LoadConfig(conf)
	if err == nil {
		t.Fatalf("Should have errored on changed listen port")
	}
}

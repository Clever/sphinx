package main

import (
	"flag"
	"github.com/Clever/sphinx/config"
	"github.com/Clever/sphinx/daemon"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	configfile = flag.String("config", "example.yaml", "/path/to/configuration.yaml")
	validate   = flag.Bool("validate", false, "Validate configuration and exit")
)

func main() {

	flag.Parse()

	conf, err := config.New(*configfile)
	if err != nil {
		log.Fatalf("LOAD_CONFIG_FAILED: %s", err.Error())
	}

	sphinxd, err := daemon.New(conf)
	if err != nil {
		log.Fatal(err)
	}

	setupSighupHandler(sphinxd, sighupHandler)
	if *validate {
		print("configuration parsed and Sphinx loaded fine. not starting dameon.")
		return
	}

	sphinxd.Start()
}

// sighupHandler is the default HUP signal handler. It is defined separately so tests can
// overwrite it with their own handler.
func sighupHandler(d daemon.Daemon) {
	conf, err := config.New(*configfile)
	if err != nil {
		log.Println("RELOAD_CONFIG_FILE_FAILED: " + err.Error())
		return
	}
	err = d.LoadConfig(conf)
	if err != nil {
		log.Println("RELOAD_CONFIG_FILE_FAILED: " + err.Error())
		return
	}
	log.Println("Reloaded config file")
}

// setupSighupHandler craetes a channel to listen for HUP signals and process them.
func setupSighupHandler(d daemon.Daemon, handler func(daemon.Daemon)) {
	sigc := make(chan os.Signal)
	signal.Notify(sigc, syscall.SIGHUP)
	go func() {
		// Listen for HUP signals "forever", calling the hup-handler each time
		// one is received.
		for {
			<-sigc
			handler(d)
		}
	}()
}

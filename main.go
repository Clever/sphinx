package main

import (
	"flag"
	"github.com/Clever/sphinx/config"
	"github.com/Clever/sphinx/daemon"
	"log"
)

var (
	configfile = flag.String("config", "example.yaml", "/path/to/configuration.yaml")
	validate   = flag.Bool("validate", false, "Validate configuration and exit")
)

func main() {

	flag.Parse()

	config, err := config.NewConfiguration(*configfile)
	if err != nil {
		log.Fatalf("LOAD_CONFIG_FAILED: %s", err.Error())
	}

	sphinxd, err := daemon.NewDaemon(config)
	if err != nil {
		log.Fatal(err)
	}

	if *validate {
		print("configuration parsed and Sphinx loaded fine. not starting dameon.")
		return
	}

	sphinxd.Start()
}

package main

import (
	"flag"
	//"fmt"
)

var (
	config = flag.String("config", "sphinx.yaml", "/path/to/configuration.yaml")
)

func main() {

	flag.Parse()
	NewSphinxDaemon(NewConfiguration(*config))
}

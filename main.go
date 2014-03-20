package main

import (
	"flag"
	"fmt"
)

var (
	config = flag.String("config", "sphinx.yaml", "/path/to/configuration.yaml")
)

func main() {
	fmt.Println(`
        Say on, sweet Sphinx! thy dirges 
        Are pleasant songs to me. 
        Deep love lieth under 
        These pictures of time; 
        They fade in the light of 
        Their meaning sublime.
          - Ralph Waldo Emerson.
          `)

	flag.Parse()
	NewConfiguration(*config)
}

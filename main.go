package sphinx

import (
	"flag"
	//"fmt"
)

var (
	config = flag.String("config", "example.yaml", "/path/to/configuration.yaml")
)

func main() {

	flag.Parse()
	NewDaemon(NewConfiguration(*config))
}

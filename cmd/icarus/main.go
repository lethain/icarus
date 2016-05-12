package main

import (
	"flag"
	"fmt"

	"github.com/lethain/icarus"
)

var loc = flag.String("loc", ":8080", "host:port to bind to, defaults to :8080")

func main() {
	flag.Parse()
	fmt.Printf("Starting Icarus on %v.\n", *loc)
	icarus.Serve(*loc)
}

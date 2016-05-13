package main

import (
	"flag"
	"fmt"

	"github.com/lethain/icarus"
)

var loc = flag.String("loc", ":8080", "host:port to bind to, defaults to :8080")
var templates = flag.String("templates", "templates/", "directory holding layouts/ and includes/ directories for templates")

func main() {
	flag.Parse()
	fmt.Printf("Starting Icarus on %v.\n", *loc)
	icarus.Serve(*loc, *templates)
}

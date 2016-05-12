package main

import (
	"fmt"
	"flag"
	
	"github.com/lethain/icarus"
)

var loc = flag.String("loc", ":8080", "host:port to bind to, defaults to :8080")


func main() {
	fmt.Printf("Starting Icarus on %v.\n", *loc)
	icarus.Serve(*loc)
}

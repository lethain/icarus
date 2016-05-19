package main

import (
	"flag"
	"log"

	"github.com/lethain/icarus"
)

var configPath = flag.String("config", "config.json", "path to configuration file, defaults to config.json")

func main() {
	flag.Parse()
	cfg, err := icarus.NewConfigFromFile(*configPath)
	if err != nil {
		log.Fatalf("error loading config: %v\n", err)
	}
	log.Printf("Starting Icarus via config: %v.\n", cfg)
	icarus.Serve(cfg)
}

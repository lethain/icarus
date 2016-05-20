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
		log.Fatalf("failed configuring redis: %v", err)
	}
	log.Printf("loaded configuration: %v", cfg)
	err = icarus.ConfigRedis(cfg)
	if err != nil {
		log.Fatalf("failed configuring redis: %v", err)
	}
	err = icarus.ConfigSearch(cfg)
	if err != nil {
		log.Fatalf("failed configuring search: %v", err)
	}	
	
	icarus.ReindexAll()
}
	

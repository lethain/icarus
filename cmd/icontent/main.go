package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/lethain/icarus"
)

var asMarkdown = flag.Bool("markdown", false, "Force evaluating as Markdown.")
var asHTML = flag.Bool("html", false, "Force evaluating as HTML.")
var configPath = flag.String("config", "config.json", "path to configuration file, defaults to config.json")

func render(filename string, content string) (*icarus.Page, error) {
	if *asMarkdown {
		return icarus.RenderMarkdown(content)
	}
	if *asHTML {
		return icarus.RenderHTML(content)
	}
	return icarus.Render(filename, content)
}

func main() {
	flag.Parse()
	files := flag.Args()

	// setup redis and search for loading files
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
	// find the files to parse, render, load and index
	if len(files) == 0 {
		log.Fatalf("must specify at least one file to load")
	}
	pages := make([]*icarus.Page, 0, len(files))
	for _, file := range files {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			log.Printf("failed to read %v: %v", file, err)
			continue
		}
		page, err := render(file, string(content))
		if err != nil {
			log.Printf("failed to render %v: %v", file, err)
			continue
		}
		pages = append(pages, page)
	}

	log.Printf("read %v pages from disk, now loading them into Icarus", len(pages))
	for _, page := range pages {
		log.Printf("synchronizing %v", page.Slug)
		err := page.Sync()
		if err != nil {
			log.Printf("failed to load %v (%v) into redis: %v", page.Title, page.Slug, err)
		}
	}
}

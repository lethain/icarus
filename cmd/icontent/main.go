package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/lethain/icarus"
)

var asMarkdown = flag.Bool("markdown", false, "Force evaluating as Markdown.")
var asHTML = flag.Bool("html", false, "Force evaluating as HTML.")

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
		err := page.Sync()
		if err != nil {
			log.Printf("failed to load %v (%v) into redis: %v", page.Title, page.Slug, err)
		}
	}
}

package main

import (
	"flag"
	"fmt"

	"github.com/lethain/icarus"
)

var loc = flag.String("loc", ":8080", "host:port to bind to, defaults to :8080")
var templates = flag.String("templates", "templates/", "directory holding layouts/ and includes/ directories for templates")
var static = flag.String("static", "static/", "path to directory holding static assets")
var domain = flag.String("domain", "http://lethain.com", "protocol and domain, e.g. http://yourthing.com")

func main() {
	flag.Parse()

	rss := icarus.RSSConfig{Path: "/feeds/", Title: "RSS Feed"}
	cfg := icarus.Config{NetLoc: *loc, DomainUrl: *domain, RSS: rss, TemplateDir: *templates, StaticDir: *static}

	fmt.Printf("Starting Icarus via config: %v.\n", cfg.String())
	icarus.Serve(&cfg)
}

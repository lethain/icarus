package icarus

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

var templateCache map[string]*template.Template

// Handle rendering generic pages stored in Redis.
func handlePage(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Path[1:]
	if strings.HasSuffix(slug, "/") {
		slug = slug[:len(slug)-1]
	}
	page, err := PageFromRedis(slug)
	if err != nil {
		fmt.Fprintf(w, "Error opening slug '%s'\n%v\n\n%v", r.URL.Path[1:], err, page)
		return
	}

	params := make(map[string]interface{})
	params["Page"] = page
	params["Now"] = time.Now()
	params["Query"] = ""
	// todo: push this into a param/config
	params["DomainUrl"] = "http://lethain.com"
	params["RSS"] = map[string]string{
		"Path":  "/feeds/",
		"Title": "Page Feed",
	}

	recent, err := RecentPages(0, 5)
	if err != nil {
		log.Printf("error generating recent pages: %v", err)
		recent = []*Page{}
	}
	params["Recent"] = recent
	trending, err := TrendingPages(0, 5)
	if err != nil {
		log.Printf("error generating trending pages: %v", err)
		trending = []*Page{}
	}
	params["Trending"] = trending

	if !page.Draft {
		previous, err := Surrounding(page, 2, true)
		if err != nil {
			log.Printf("error generating previous pages: %v", err)
			previous = []*Page{}
		}
		params["Previous"] = previous
		following, err := Surrounding(page, 2, false)
		if err != nil {
			log.Printf("error generating following pages: %v", err)
			following = []*Page{}
		}
		params["Following"] = following
	} else {
		params["Previous"] = []*Page{}
		params["Following"] = []*Page{}
	}

	params["Similar"] = []*Page{}


	err = renderTemplate(w, "page.html", params)
	if err != nil {
		log.Printf("error rendering: %v", err)
	}
	err = Track(page, r)
	if err != nil {
		log.Printf("error tracking %v: %v", page.Slug, err)
	}
}

func Serve(loc string, templatePath string, staticPath string) {
	loadTemplates(templatePath)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))
	http.HandleFunc("/", handlePage)
	http.ListenAndServe(loc, nil)
}

func renderTemplate(w http.ResponseWriter, name string, data map[string]interface{}) error {
	tmpl, ok := templateCache[name]
	if !ok {
		return fmt.Errorf("template %s does not exist", name)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.ExecuteTemplate(w, "base", data)
}

func loadTemplates(templatePath string) {
	if templateCache == nil {
		templateCache = make(map[string]*template.Template)
	}

	layouts, err := filepath.Glob(templatePath + "layouts/*.html")
	if err != nil {
		log.Fatal(err)
	}

	includes, err := filepath.Glob(templatePath + "includes/*.html")
	if err != nil {
		log.Fatal(err)
	}

	for _, layout := range layouts {
		files := append(includes, layout)
		log.Printf("loading and composing templates for %v : %v\n", filepath.Base(layout), files)

		templateCache[filepath.Base(layout)] = template.Must(template.ParseFiles(files...))
	}
}

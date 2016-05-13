package icarus

import (
	"fmt"
	"path/filepath"
	"log"
	"net/http"
	"text/template"
)

var templateCache map[string]*template.Template

// Handle rendering generic pages stored in Redis.
func handlePage(w http.ResponseWriter, r *http.Request) {
	page, err := PageFromRedis(r.URL.Path[1:])
	if err != nil {
		fmt.Fprintf(w, "Error opening slug '%s'\n%v\n\n%v", r.URL.Path[1:], err, page)
		return
	}

	params := make(map[string]interface{})
	params["Page"] = page
	err = renderTemplate(w, "page.html", params)
	if err != nil {
		log.Printf("error rendering: %v", err)
	}
}

func Serve(loc string, templatePath string) {
	loadTemplates(templatePath)
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

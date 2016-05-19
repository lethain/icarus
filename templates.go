// Handling templates and such.
package icarus

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"text/template"
)

var templateCache map[string]*template.Template

func renderTemplate(w http.ResponseWriter, name string, data map[string]interface{}) error {
	tmpl, ok := templateCache[name]
	if !ok {
		return fmt.Errorf("template %s does not exist", name)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.ExecuteTemplate(w, "base", data)
}

func ConfigTemplates(cfg *Config) error {
	templatePath := cfg.Blog.TemplateDir

	if templateCache == nil {
		templateCache = make(map[string]*template.Template)
	}

	layouts, err := filepath.Glob(templatePath + "layouts/*.html")
	if err != nil {
		return err
	}

	includes, err := filepath.Glob(templatePath + "includes/*.html")
	if err != nil {
		return err
	}
	for _, layout := range layouts {
		files := append(includes, layout)
		log.Printf("loading and composing templates for %v : %v\n", filepath.Base(layout), files)
		templateCache[filepath.Base(layout)] = template.Must(template.ParseFiles(files...))
	}
	return nil
}

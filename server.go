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

func buildSidebar(cfg *Config, p *Page) (map[string]interface{}, error) {
	params := make(map[string]interface{})

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
	if p != nil && !p.Draft {
		previous, err := Surrounding(p, 2, true)
		if err != nil {
			log.Printf("error generating previous pages: %v", err)
			previous = []*Page{}
		}
		params["Previous"] = previous

		following, err := Surrounding(p, 2, false)
		if err != nil {
			log.Printf("error generating following pages: %v", err)
			following = []*Page{}
		}
		params["Following"] = following

		// TOOD: implement similar
		params["Similar"] = []*Page{}
	} else {
		params["Previous"] = []*Page{}
		params["Following"] = []*Page{}
		params["Similar"] = []*Page{}
	}
	return params, nil
}

func defaultParams(cfg *Config, p *Page) (map[string]interface{}, error) {
	params, err := buildSidebar(cfg, p)
	if err != nil {
		return params, err
	}
	params["Cfg"] = cfg
	params["Page"] = p
	params["Now"] = time.Now()
	params["Query"] = ""
	return params, nil
}

// build http.HandlerFunc for rendering generic pages stored in Redis.
func makePageHandler(cfg *Config) http.HandlerFunc {
	handle := func(w http.ResponseWriter, r *http.Request) {
		slug := r.URL.Path[1:]
		if strings.HasSuffix(slug, "/") {
			slug = slug[:len(slug)-1]
		}
		p, err := PageFromRedis(slug)
		if err != nil {
			errorPage(w, r, p, err)
			return
		}

		params, err := defaultParams(cfg, p)
		if err != nil {
			errorPage(w, r, p, err)
			return
		}

		err = renderTemplate(w, "page.html", params)
		if err != nil {
			errorPage(w, r, p, err)
			return
		}

		err = Track(p, r)
		if err != nil {
			errorPage(w, r, p, err)
			return
		}
	}
	return handle
}

func errorPage(w http.ResponseWriter, r *http.Request, p *Page, err error) {
	log.Printf("Error opening slug '%s'\n%v\n\n%v", r.URL.Path[1:], err, p)
	fmt.Fprintf(w, "Error opening slug '%s'\n%v\n\n%v", r.URL.Path[1:], err, p)
}

func Serve(cfg *Config) {
	loadTemplates(cfg.TemplateDir)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(cfg.StaticDir))))
	http.HandleFunc("/", makePageHandler(cfg))
	http.ListenAndServe(cfg.NetLoc, nil)
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

package icarus

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
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

func defaultParams(cfg *Config, p *Page, r *http.Request) (map[string]interface{}, error) {
	params, err := buildSidebar(cfg, p)
	if err != nil {
		return params, err
	}
	params["Cfg"] = cfg
	params["Page"] = p
	params["Path"] = r.URL.Path[1:]
	params["Now"] = time.Now()
	params["Query"] = ""
	return params, nil
}

func makeTagHandler(cfg *Config) http.HandlerFunc {
	handle := func(w http.ResponseWriter, r *http.Request) {
		tag := getSlug(r)[5:]
		list := fmt.Sprintf(TagPagesZsetByTrend, tag)
		tagHandler := makeListHandler(cfg, list, fmt.Sprintf("Pages for %v Tag", tag))
		tagHandler(w, r)
	}
	return handle
}

func makeTagsHandler(cfg *Config, title string) http.HandlerFunc {
	tagHandler := makeTagHandler(cfg)

	handle := func(w http.ResponseWriter, r *http.Request) {
		// this matches /tags/ but also the tag detail page
		// at /tags/etc , so disambiguate between those
		// pages here
		slug := getSlug(r)
		if slug != "tags" {
			tagHandler(w, r)
			return
		}
		params, err := defaultParams(cfg, nil, r)
		if err != nil {
			errorPage(w, r, cfg, nil, err)
			return
		}
		params["Title"] = title
		allTags, err := GetAllTags()
		if err != nil {
			errorPage(w, r, cfg, nil, err)
			return
		}
		params["Tags"] = allTags
		err = renderTemplate(w, "tags.html", params)
		if err != nil {
			errorPage(w, r, cfg, nil, err)
			return
		}

	}
	return handle

}

func makeListHandler(cfg *Config, list string, title string) http.HandlerFunc {
	handle := func(w http.ResponseWriter, r *http.Request) {
		offset := 0
		offsetStr := r.URL.Query().Get("offset")
		if offsetStr != "" {
			o, err := strconv.ParseInt(offsetStr, 10, 32)
			if err != nil {
				log.Printf("error parsing offset: %v", err)
			} else {
				offset = int(o)
			}
		}
		total, err := PagesInList(list)
		if err != nil {
			errorPage(w, r, cfg, nil, err)
			return
		}
		pgs, err := PagesForList(list, offset, cfg.ListCount, true)
		if err != nil {
			errorPage(w, r, cfg, nil, err)
			return
		}
		params, err := defaultParams(cfg, nil, r)
		if err != nil {
			errorPage(w, r, cfg, nil, err)
			return
		}
		params["Title"] = title
		params["Pages"] = pgs
		params["Paginator"] = NewPaginator(offset, total, cfg.ListCount, 10)
		err = renderTemplate(w, "list.html", params)
		if err != nil {
			errorPage(w, r, cfg, nil, err)
			return
		}

	}
	return handle
}

// Determine the slug for a request, including defaulting to
// the latest post if no slug is specified.
func getSlug(r *http.Request) string {
	slug := r.URL.EscapedPath()[1:]
	if strings.HasSuffix(slug, "/") {
		slug = slug[:len(slug)-1]
	}
	return slug
}

// build http.HandlerFunc for rendering generic pages stored in Redis.
func makePageHandler(cfg *Config, indexHandler http.HandlerFunc) http.HandlerFunc {
	handle := func(w http.ResponseWriter, r *http.Request) {
		slug := getSlug(r)
		if slug == "" {
			indexHandler(w, r)
			return
		}
		p, err := PageFromRedis(slug)
		if err != nil {
			if _, ok := err.(*NoSuchPagesError); ok {
				notFoundPage(w, r, cfg, err)
			} else {
				errorPage(w, r, cfg, p, err)
			}
			return
		}
		params, err := defaultParams(cfg, p, r)
		if err != nil {
			errorPage(w, r, cfg, p, err)
			return
		}
		err = renderTemplate(w, "page.html", params)
		if err != nil {
			errorPage(w, r, cfg, p, err)
			return
		}
		err = Track(p, r)
		if err != nil {
			errorPage(w, r, cfg, p, err)
			return
		}
	}
	return handle
}

func errorPage(w http.ResponseWriter, r *http.Request, cfg *Config, p *Page, err error) {
	params := map[string]interface{}{
		"Title":     "500",
		"Previous":  []*Page{},
		"Following": []*Page{},
		"Similar":   []*Page{},
		"Recent":    []*Page{},
		"Trending":  []*Page{},
		"Cfg":       cfg,
		"Path":      r.URL.Path[1:],
		"Query":     "",
	}

	log.Printf("Error opening slug '%s'\n%v\n\n%v", r.URL.Path[1:], err, p)
	err = renderTemplate(w, "500.html", params)
	if err != nil {
		fmt.Fprint(w, "Things are going very poorly. Please check back later.")
	}
}

func notFoundPage(w http.ResponseWriter, r *http.Request, cfg *Config, err error) {
	params, err := defaultParams(cfg, nil, r)
	if err != nil {
		errorPage(w, r, cfg, nil, err)
		return
	}
	err = renderTemplate(w, "404.html", params)
	if err != nil {
		errorPage(w, r, cfg, nil, err)
		return
	}

}

func Serve(cfg *Config) {
	loadTemplates(cfg.TemplateDir)
	recentHandler := makeListHandler(cfg, PageZsetByTime, "Recent Pages")

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(cfg.StaticDir))))
	http.HandleFunc("/list/trending/", makeListHandler(cfg, PageZsetByTrend, "Popular Pages"))
	http.HandleFunc("/list/recent/", recentHandler)
	http.HandleFunc("/tags/", makeTagsHandler(cfg, "Tags By Page Count"))
	http.HandleFunc("/", makePageHandler(cfg, recentHandler))
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

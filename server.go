package icarus

import (
	"fmt"
	"log"
	"net/http"
	"html"

	"strconv"
	"strings"

	"time"
)

// TODO: move this to Config
const PagesInModules = 3

func buildSidebar(cfg *Config, p *Page) (map[string]interface{}, error) {
	params := make(map[string]interface{})

	recent, err := RecentPages(0, PagesInModules)
	if err != nil {
		log.Printf("error generating recent pages: %v", err)
		recent = []*Page{}
	}
	params["Recent"] = recent
	trending, err := TrendingPages(0, PagesInModules)
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

		similar, err := SimilarPages(p, 0, PagesInModules)
		if err != nil {
			log.Printf("error generating similar pages: %v", err)
			similar = []*Page{}
		}
		params["Similar"] = similar
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

func makeFeedsHandler(cfg *Config) http.HandlerFunc {
	handle := func(w http.ResponseWriter, r *http.Request) {
		pgs, err := PagesForList(PageZsetByTime, 0, cfg.Blog.ResultsPerPage, true)
		if err != nil {
			errorPage(w, r, cfg, nil, err)
			return
		}
		feed, err := BuildAtomFeed(cfg, pgs)
		if err != nil {
			errorPage(w, r, cfg, nil, err)
			return
		}
		atom, err := feed.ToAtom()
		if err != nil {
			errorPage(w, r, cfg, nil, err)
			return
		}
		fmt.Fprint(w, atom)
	}
	return handle
}

func makeSearchHandler(cfg *Config) http.HandlerFunc {
	handle := func(w http.ResponseWriter, r *http.Request) {
		qu := r.URL.Query().Get("q")
		q := html.EscapeString(qu)

		slugs, err := Search(q)
		if err != nil {
			errorPage(w, r, cfg, nil, err)
			return
		}

		pgs, err := PagesFromRedis(slugs)
		if err != nil {
			errorPage(w, r, cfg, nil, err)
			return
		}

		params, err := defaultParams(cfg, nil, r)
		if err != nil {
			errorPage(w, r, cfg, nil, err)
			return
		}
		params["Query"] = q
		params["Title"] = fmt.Sprintf("Results for \"%v\"", q)
		params["Pages"] = pgs
		err = renderTemplate(w, "list.html", params)
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
		pgs, err := PagesForList(list, offset, cfg.Blog.ResultsPerPage, true)
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
		params["Paginator"] = NewPaginator(offset, total, cfg.Blog.ResultsPerPage, cfg.Blog.PagesInPaginator)
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

	slug := ""
	if p != nil {
		slug = p.Slug
	}

	log.Printf("Error opening slug '%s'\n%v\n\n%v", r.URL.Path[1:], err, slug)
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
	err := ConfigTemplates(cfg)
	if err != nil {
		log.Fatalf("failed configuring templates: %v", err)
	}
	err = ConfigRedis(cfg)
	if err != nil {
		log.Fatalf("failed configuring redis: %v", err)
	}
	err = ConfigSearch(cfg)
	if err != nil {
		log.Fatalf("failed configuring search: %v", err)
	}

	recentHandler := makeListHandler(cfg, PageZsetByTime, "Recent Pages")

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(cfg.Blog.StaticDir))))
	http.HandleFunc("/list/trending/", makeListHandler(cfg, PageZsetByTrend, "Popular Pages"))
	http.HandleFunc("/list/recent/", recentHandler)
	http.HandleFunc("/tags/", makeTagsHandler(cfg, "Tags By Page Count"))
	http.HandleFunc("/feeds/", makeFeedsHandler(cfg))
	http.HandleFunc("/search/", makeSearchHandler(cfg))
	http.HandleFunc("/", makePageHandler(cfg, recentHandler))
	http.ListenAndServe(cfg.Server.Loc, nil)
}

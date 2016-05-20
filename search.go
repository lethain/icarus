package icarus

import (
	"errors"
	"log"

	"github.com/blevesearch/bleve"
)

var searchDir = "searchIndex/"
var searchIndex bleve.Index

func ConfigSearch(cfg *Config) error {
	// TODO: override searchdir from config
	idx, err := bleve.Open(searchDir)
	if err != nil {
		mapping := bleve.NewIndexMapping()
		idx, err = bleve.New(searchDir, mapping)
		if err != nil {
			return err
		}
	}
	searchIndex = idx
	return nil
}

func Search(qs string) ([]string, error) {
	if searchIndex == nil {
		return []string{}, errors.New("shared searchIndex is not initialized")
	}
	q := bleve.NewMatchQuery(qs)
	sr := bleve.NewSearchRequest(q)
	res, err := searchIndex.Search(sr)
	if err != nil {
		return []string{}, err
	}
	slugs := []string{}
	for _, hit := range res.Hits {
		slugs = append(slugs, hit.ID)
	}
	return slugs, nil
}

func ReindexAll() error {
	count := 10
	list := PageZsetByTime
	numPages, err := PagesInList(list)
	if err != nil {
		return err
	}

	for i := 0; i < numPages; i += 10 {
		pgs, err := PagesForList(list, i, count, true)
		if err != nil {
			return err
		}
		if len(pgs) == 0 {
			break
		}
		log.Printf("indexing pages %v to %v: %v", i, i+count, pgs[0])
		err = IndexPages(pgs)
		if err != nil {
			return err
		}
	}
	return nil
}

func IndexPage(p *Page) error {
	return IndexPages([]*Page{p})
}

func IndexPages(pgs []*Page) error {
	if searchIndex == nil {
		return errors.New("shared searchIndex is not initialized")
	}
	for _, p := range pgs {
		err := searchIndex.Index(p.Slug, p)
		if err != nil {
			return err
		}
	}
	return nil
}

func UnindexPage(p *Page) error {
	// TODO: implement
	return nil

}

package icarus

import (
	"errors"

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

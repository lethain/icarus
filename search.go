package icarus

import (
	"github.com/blevesearch/bleve"
)


var searchDir = "searchIndex/"
var searchIndex bleve.Index


func ConfigureSearch(cfg *Config) error {
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



func IndexPage(p *Page) error {
	return IndexPages([]*Page{p})
}

func IndexPages(pgs []*Page) error {
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

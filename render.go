package icarus

import (
	"fmt"
	"path/filepath"
)



func Render(filename string, contents string) (*Page, error) {
	switch filepath.Ext(filename) {
	case ".md":
		return RenderMarkdown(contents)
	case ".html":
		return RenderHTML(contents)
	}
	return nil, fmt.Errorf("filename %v doesn't match any known suffixes", filename)
}

func RenderMarkdown(contents string) (*Page, error) {
	//page['html'] = markdown.markdown(page['html'], ['codehilite(css_class=highlight)', 'headerid','toc'])
	return nil, fmt.Errorf("render markdown not implemented")	
}

func RenderHTML(contents string) (*Page, error) {
	return nil, fmt.Errorf("render html not implemented")

}

func ReadHeaders(contents string) (*Page, error) {
	// we want to read off the pseudo-json headers from the
	// top of the file, and return a page object

	// default pub_date to now
	// default tags to []
	//
	p := &Page{}
	return p, nil
}

package icarus

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday"
)

func Render(filename string, content string) (*Page, error) {
	switch filepath.Ext(filename) {
	case ".md":
		return RenderMarkdown(content)
	case ".html":
		return RenderHTML(content)
	}
	return nil, fmt.Errorf("filename %v doesn't match any known suffixes", filename)
}

func RenderMarkdown(content string) (*Page, error) {
	//page['html'] = markdown.markdown(page['html'], ['codehilite(css_class=highlight)', 'headerid','toc'])
	p, content, err := ReadHeaders(content)
	if err != nil {
		return nil, fmt.Errorf("error reading headers: %v", err)
	}
	if p.Slug == "" {
		return p, fmt.Errorf("skipping %v because it has no slug", p.Title)
	}
	p.Content = string(blackfriday.MarkdownCommon([]byte(content)))
	return p, nil
}

func RenderHTML(content string) (*Page, error) {
	p, content, err := ReadHeaders(content)
	if err != nil {
		fmt.Errorf("error reading headers: %v", err)
	}
	p.Content = content
	return p, nil
}

/*
Create a new Page with the headers stripped out of it.

Headers look like this:

"tags": ["python", "redis"],
"title": "Storing Bounded Timeboxes in Redis",
"summary": "This is the summary...",
"slug": "storing-bounded-timeboxes-in-redis",

Including a \n\n (outside of string) to designate the end of the section.
In general, you can describe the header as a JSON dictionary without the
opening or closing {}s.
*/
func ReadHeaders(content string) (*Page, string, error) {
	head, rest, err := splitHeader(content)
	if err != nil {
		return nil, "", err
	}
	if strings.HasSuffix(head, "\n") {
		head = head[:len(head)-1]
	}
	if strings.HasSuffix(head, ",") {
		head = head[:len(head)-1]
	}
	headers := "{\n" + head + "}\n"

	p := &Page{}
	if err := json.Unmarshal([]byte(headers), &p); err != nil {
		return p, rest, err
	}
	p.EnsureEditDate()
	p.EnsurePubDate()

	// default pub_date to now
	return p, rest, nil
}

func splitHeader(content string) (string, string, error) {
	stacks := make(map[rune]int)
	var prev rune
	for i, c := range content {
		if c == '\n' && prev == '\n' && (stacks['"']%2 == 0) && (stacks['['] == 0) && (stacks['{'] == 0) {
			return content[:i], content[i:], nil
		}
		// ignore escaped things
		if prev == '\\' {
			prev = rune(0)
			continue
		}

		// make sure we're not in a quote, list or whatever
		switch c {
		case '[', '{', '"':
			stacks[c] += 1
		case ']':
			stacks['['] -= 1
		case '}':
			stacks['['] -= 1
		}
		prev = c
	}
	return "", "", fmt.Errorf("couldn't find header split")
}

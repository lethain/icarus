package icarus

import (
	"fmt"
	"time"
)

func PageFromRedis(slug string) (*Page, error) {
	p := &Page{}
	rc, err := GetConfiguredRedisClient()
	if err != nil {
		return p, fmt.Errorf("failed to connect to redis: %v", err)
	}
	raw, err :=rc.Cmd("GET", "page."+slug).Str()
	if err != nil {
		return p, fmt.Errorf("failed retrieving slug %v from redis: %v", slug, err)		
	}
	
	fmt.Printf("raw: %v\n", raw)
	p.Content = raw
	return p, nil
}

type Page struct {
	Slug string
	Tags []string
	Title string
	Summary string
	Content string
	Published bool
	pubDate int32 `json:"pub_date"`
	editDate int32 `json:"edit_date"`
}

func (p *Page) PubDate() time.Time {
	return time.Unix(int64(p.pubDate), 0)
}

func (p *Page) EditDate() time.Time {
	return time.Unix(int64(p.pubDate), 0)
}



type Pages []Page

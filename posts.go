package icarus

import (
	"fmt"
	"log"
	"time"
	"encoding/json"
)

// Retrieve a list of slugs from Redis.
func PagesFromRedis(slugs []string) ([]*Page, error) {
	pages := make([]*Page, 0, len(slugs))
	keys := make([]string, 0, len(slugs))
	for _, slug := range slugs {
		p := &Page{Slug: slug}
		pages = append(pages, p)
		keys = append(keys, p.Key())
	}
	rc, err := GetConfiguredRedisClient()
	if err != nil {
		return pages, fmt.Errorf("failed to connect to redis: %v", err)
	}
	raws, err := rc.Cmd("MGET", keys).List()
	if err != nil {
		return pages, fmt.Errorf("failed retrieving slugs %v from redis: %v", slugs, err)
	}

	for i, raw := range raws {
		if err := json.Unmarshal([]byte(raw), pages[i]); err != nil {
			return pages, err
		}
	}
	return pages, nil
}

// Retrieve one page from Redis.
func PageFromRedis(slug string) (*Page, error) {
	pages, err := PagesFromRedis([]string{slug})
	if err != nil {
		return nil, err
	}
	if len(pages) != 1 {
		return nil, fmt.Errorf("retrieve none-one number of values for %v", slug)
	}
	return pages[0], nil
}


type Page struct {
	Slug      string `json:"slug"`
	Tags      []string `json:"tags"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	Content   string `json:"content"`
	Draft     bool   `json:"draft"`
	PubDateStr   string `json:"pub_date"`
	EditDateStr  string `json:"edit_date"`
}

// Generate the Redis key for this page.
func (p *Page) Key() string {
	return "page." + p.Slug
}

// Synchronize this page to Redis.
func (p *Page) Sync() error {
	rc, err := GetConfiguredRedisClient()
	if err != nil {
		return fmt.Errorf("failed connecting to redis: %v", err)
	}
	asJSON, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed marshalling page %v: %v", p.Slug, err)
	}
	_, err = rc.Cmd("SET", p.Key(), asJSON).Str()
	return nil
}

func (p *Page) getDate(date string) time.Time {
	t, err := time.Parse(time.RFC850, date)
	if err != nil {
		log.Printf("failed to parse %v for getDate", date)
		return time.Time{}
	}
	return t
}

func (p *Page) PubDate() time.Time {
	return p.getDate(p.PubDateStr)
}

func (p *Page) EditDate() time.Time {
	return p.getDate(p.EditDateStr)
}

func (p *Page) InitPubDate() {
	p.PubDateStr = time.Now().Format(time.RFC850)
}

func (p *Page) InitEditDate() {
	p.EditDateStr = time.Now().Format(time.RFC850)
}

func (p *Page) EnsurePubDate() {
	if p.PubDateStr == "" {
		p.InitPubDate()
	}
}

func (p *Page) EnsureEditDate() {
	if p.EditDateStr == "" {
		p.InitEditDate()
	}

}

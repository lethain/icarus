package icarus

import (
	"fmt"
	"log"
	"time"

	"encoding/json"
)

func PageFromRedis(slug string) (*Page, error) {
	p := &Page{Slug: slug}
	rc, err := GetConfiguredRedisClient()
	if err != nil {
		return p, fmt.Errorf("failed to connect to redis: %v", err)
	}
	raw, err := rc.Cmd("GET", p.Key()).Str()
	if err != nil {
		return p, fmt.Errorf("failed retrieving slug %v from redis: %v", slug, err)
	}
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		return p, err
	}
	return p, nil
}

type Page struct {
	Slug      string `json:"slug"`
	Tags      []string `json:"tags"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	Content   string `json:"content"`
	Published bool `json:"published"`
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

type Pages []Page

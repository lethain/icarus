package icarus

import (
	"encoding/json"
	"fmt"
	"time"
)

type NoSuchPagesError struct {
	msg   string
	Slugs []string
}

func (e *NoSuchPagesError) Error() string {
	return e.msg
}

// Retrieve a list of slugs from Redis.
func PagesFromRedis(slugs []string) ([]*Page, error) {
	pages := make([]*Page, 0, len(slugs))
	keys := make([]string, 0, len(slugs))
	for _, slug := range slugs {
		p := &Page{Slug: slug}
		pages = append(pages, p)
		keys = append(keys, p.Key())
	}
	rc, err := GetRedisClient()
	defer PutRedisClient(rc)

	if err != nil {
		return pages, fmt.Errorf("failed to connect to redis: %v", err)
	}
	if len(keys) == 0 {
		return []*Page{}, nil
	}

	raws, err := rc.Cmd("MGET", keys).List()

	nonEmpty := 0
	for _, raw := range raws {
		if raw != "" {
			nonEmpty += 1
		}
	}
	if err != nil || nonEmpty == 0 {
		msg := fmt.Sprintf("failed retrieving slugs %v from redis: %v", slugs, err)
		return pages, &NoSuchPagesError{msg, slugs}
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
	Slug        string   `json:"slug"`
	Tags        []string `json:"tags"`
	Title       string   `json:"title"`
	Summary     string   `json:"summary"`
	Content     string   `json:"html"`
	Draft       bool     `json:"draft"`
	PubDateStr  int64    `json:"pub_date"`
	EditDateStr int64    `json:"edit_date"`
}

// Generate the Redis key for this page.
func (p *Page) Key() string {
	return "page." + p.Slug
}

// Synchronize this page to Redis.
func (p *Page) Sync() error {
	rc, err := GetRedisClient()
	defer PutRedisClient(rc)

	if err != nil {
		return fmt.Errorf("failed connecting to redis: %v", err)
	}
	asJSON, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed marshalling page %v: %v", p.Slug, err)
	}
	_, err = rc.Cmd("SET", p.Key(), asJSON).Str()
	if err != nil {
		return err
	}
	if !p.Draft {
		err := RegisterPage(p)
		if err != nil {
			return err
		}
		err = IndexPage(p)
		if err != nil {
			return err
		}

	} else {
		err := UnregisterPage(p)
		if err != nil {
			return err
		}
		err = UnindexPage(p)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Page) getDate(date int64) time.Time {
	t := time.Unix(date, 0)
	return t
}

func (p *Page) PubDate() time.Time {
	return p.getDate(p.PubDateStr)
}

func (p *Page) EditDate() time.Time {
	return p.getDate(p.EditDateStr)
}

func (p *Page) InitPubDate() {
	p.PubDateStr = CurrentTimestamp()
}

func (p *Page) InitEditDate() {
	p.EditDateStr = CurrentTimestamp()
}

func (p *Page) EnsurePubDate() {
	if p.PubDateStr == 0 {
		p.InitPubDate()
	}
}

func (p *Page) EnsureEditDate() {
	if p.EditDateStr == 0 {
		p.InitEditDate()
	}
}

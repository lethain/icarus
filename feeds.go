package icarus

import (
	"time"
	"github.com/gorilla/feeds"
)

func BuildAtomFeed(cfg *Config, ps []*Page) (*feeds.Feed, error) {
	now := time.Now()
	feed := &feeds.Feed{
		Title: cfg.Blog.Name,
		Link: &feeds.Link{Href:cfg.BaseURL()},
		Created: now,
		Items: []*feeds.Item{},
	}

	for _, p := range ps {
		feed.Items = append(feed.Items, &feeds.Item{
			Title: p.Title,
			Description: p.Content,
			Created: p.PubDate(),
			Link: &feeds.Link{Href:cfg.BaseURL() + p.Slug + "/"},
		})
	}
	return feed, nil
}

package icarus

import (
	"time"
	"github.com/gorilla/feeds"
)

func BuildAtomFeed(cfg *Config, ps []*Page) (*feeds.Feed, error) {
	now := time.Now()
	feed := &feeds.Feed{
		Title: cfg.BlogName,
		Link: &feeds.Link{Href:cfg.DomainUrl + "/"},
		Created: now,
		Items: []*feeds.Item{},
	}

	for _, p := range ps {
		feed.Items = append(feed.Items, &feeds.Item{
			Title: p.Title,
			Description: p.Content,
			Created: p.PubDate(),
			Link: &feeds.Link{Href:cfg.DomainUrl + "/" + p.Slug + "/"},			
		})
	}
	return feed, nil
}

// Configuration for Icarus instance.
package icarus

import (
	"fmt"
)



type RSSConfig struct {
	Path string
	Title string
}

func (r *RSSConfig) String() string {
	return fmt.Sprintf("RSSConfig(%v, %v)", r.Path, r.Title)
}

type Config struct {
	NetLoc    string
	DomainUrl string
	RSS       RSSConfig
	TemplateDir string
	StaticDir string
}

func (c *Config) String() string {
	return fmt.Sprintf("Config(NetLoc: %v, DomainUrl: %v, TemplateDir: %v, StaticDir: %v, RSS: %v",
		c.NetLoc, c.DomainUrl, c.TemplateDir, c.StaticDir, c.RSS.String())
}

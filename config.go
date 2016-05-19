// Configuration for Icarus instance.
package icarus

import (
	"fmt"
	"io/ioutil"
	"encoding/json"
)

type RSSConfig struct {
	Path  string
	Title string
}

// replacing:
// 	NetLoc      string `json:"netloc"`
//	DomainUrl   string `json:"domain"`
type ServerConfig struct {
	Loc string
	Proto string
	Domain string
}

// .BlogName -> .Blog.Name
// .ListCount -> .Blog.ResultsPerPage
// .NumPages -> .Blog.PagesInPaginator
type BlogConfig struct {
	Name             string
	ResultsPerPage   int `json:"results_per_page"`
	PagesInPaginator int `json:"pages_in_paginator"`
	TemplateDir      string
	StaticDir        string
}

type RedisConfig struct {
	Loc string
	Proto string
	PoolSize int
}

type Config struct {
	Server ServerConfig
	RSS    RSSConfig
	Blog   BlogConfig
	Redis  RedisConfig
}

func (cfg *Config) BaseURL() string {
	return fmt.Sprintf("%v://%v/", cfg.Server.Proto, cfg.Server.Domain)

}

// Build a new configuration file from disk.
func NewConfigFromFile(path string) (*Config, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = json.Unmarshal(file, &cfg)
	return &cfg, err
}

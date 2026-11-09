package main

import (
	"fmt"
	"regexp"
)

var linkRegex = regexp.MustCompile(`href="([^"]+)"`)

type Crawler struct {
    seedURL  string
    visited  map[string]bool
    maxPages int
}

func NewCrawler(seed string, max int) *Crawler {
    return &Crawler{seedURL: seed, visited: make(map[string]bool), maxPages: max}
}

func (c *Crawler) Crawl() {
    queue := []string{c.seedURL}
    // ... implement the sequential BFS logic here ...
    fmt.Println("Sequential crawl finished.")
}

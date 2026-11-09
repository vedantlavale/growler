package main

import (
	"context"
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

func (c *Crawler) Crawl(ctx context.Context) {
    queue := []string{c.seedURL}
    // we are using context here because we want to be able to cancel the crawl, suppopse user hits the ctrl c then it should cancel all the process. context is used to stop the crawl gracefully.
    // ... implement the sequential BFS logic here ...

    fmt.Println("queue:", queue)
    fmt.Println("Sequential crawl finished.")
}

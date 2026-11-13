package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
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

func fetch(ctx context.Context, url string) (string, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return "", err
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
	}
	return string(body), nil
}

func (c *Crawler) Crawl(ctx context.Context) {

	body, err := fetch(ctx, c.seedURL)
	if err != nil {
		fmt.Println("fetch error:", err)
		return
	}
	fmt.Println("body bytes:", len(body))

	//fmt.Println(body)


    queue := []string{c.seedURL}
    // we are using context here because we want to be able to cancel the crawl, suppopse user hits the ctrl c then it should cancel all the process. context is used to stop the crawl gracefully.
    // ... implement the sequential BFS logic here ...

    fmt.Println("queue:", queue)
    fmt.Println("Sequential crawl finished.")
}

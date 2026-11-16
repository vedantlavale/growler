package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

var linkRegex = regexp.MustCompile(`href="([^"]+)"`)

type Crawler struct {
    seedURL  string
    visited  map[string]bool
    maxPages int
    jobs chan string
}

func NewCrawler(seed string, max int) *Crawler {
    return &Crawler{
        seedURL:  seed,
        visited:  make(map[string]bool),
        maxPages: max,
        jobs:     make(chan string),
    }
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

func normalizeURL(raw string) (string, error){
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)

	parsed.Host = strings.ToLower(parsed.Host)

	parsed.Fragment = ""
	if parsed.Path == ""{
		parsed.Path = "/"
	}

	return parsed.String(), nil
}

func extractLinks(baseURL, html string)([]string, error){
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	matches := linkRegex.FindAllStringSubmatch(html, -1)
	var links []string

	for _, match := range matches {
		href := match[1]

			ref, err := url.Parse(href)
			if err != nil {
				continue
			}

			abs := base.ResolveReference(ref)

			if abs.Scheme != "http" && abs.Scheme != "https" {
				continue
			}
			if abs.Host == "" {
				continue
			}

			links = append(links, abs.String())
	}
	return links, nil
}

func (c *Crawler) markVisited(u string) bool{
    if c.visited[u] {
        return false
    }
    c.visited[u] = true
    return true
}

func (c *Crawler) worker(ctx context.Context) {
    for {
        url, ok := <- c.jobs
        if !ok {
            return
        }
        if ctx.Err() != nil {
            return
        }
        fmt.Println("crawling URL: ", url)

        body, err := fetch(ctx, url)
        if err != nil {
            fmt.Println("fetch error:", err)
            continue
        }
        links, err := extractLinks(url, body)
        if err != nil {
            fmt.Println("extractLinks error:", err)
            continue
        }
        fmt.Println("links found: ",len(links))
    }
}

func (c *Crawler) Crawl(ctx context.Context) {

	//body, err := fetch(ctx, c.seedURL)
	//if err != nil {
	//	fmt.Println("fetch error:", err)
	//	return
	//}
	//fmt.Println("body bytes:", len(body))

	//links, err := extractLinks(c.seedURL, body)
	//if err != nil {
	//	fmt.Println("extractLinks error:", err)
	//	return
	//}

	seed, err := normalizeURL(c.seedURL)
	if err != nil {
		fmt.Println("normalizeURL error:", err)
		return
	}
	c.markVisited(seed)

    var wg sync.WaitGroup

    wg.Go(func(){
        c.worker(ctx)
    })
    c.jobs <- seed
    close(c.jobs)
    wg.Wait()

    // we are using context here because we want to be able to cancel the crawl, suppopse user hits the ctrl c then it should cancel all the process. context is used to stop the crawl gracefully.
    // ... implement the sequential BFS logic here ...

    //fmt.Println("queue:", queue)
    fmt.Println("Sequential crawl finished.")
}

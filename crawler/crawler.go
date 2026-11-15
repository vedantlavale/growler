package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
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
    queue := []string{seed}

    crawled := 0



    for len(queue) > 0 && crawled < c.maxPages {
        if ctx.Err() != nil { fmt.Println("crawl canceled"); break }

        url := queue[0]
        queue = queue[1:]

        url, err := normalizeURL(url)
        if err != nil {
            fmt.Println("normalizeURL error:", err)
            return
        }

        if c.visited[url] {
            continue
        }
        c.visited[url] = true
        crawled++

        fmt.Println("crawling URL: ",url)

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
        for _, link := range links {
        	if c.visited[link] {
        		continue
        	}
            queue = append(queue, link)
        }
    }

    // we are using context here because we want to be able to cancel the crawl, suppopse user hits the ctrl c then it should cancel all the process. context is used to stop the crawl gracefully.
    // ... implement the sequential BFS logic here ...

    //fmt.Println("queue:", queue)
    fmt.Println("crawled:", crawled)
    fmt.Println("Sequential crawl finished.")
}

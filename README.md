# growler

A small web crawler written in Go. Give it a starting URL and it walks the links it finds, page by page, until it hits the limit you set or you hit Ctrl+C.

It is also my learning project: a first Go program built to understand `net/http`, HTML parsing, goroutines, channels and context cancellation. The code is deliberately small and heavily commented so it can be read top to bottom.

## What it does right now

- Fetches a seed URL over HTTP and pulls every `href` out of the HTML.
- Resolves relative links (`/about`, `../docs`) into absolute URLs.
- Ignores anything that isn't `http` or `https`, so `mailto:` and `javascript:` links are skipped.
- Normalizes URLs before crawling (lowercases the scheme and host, drops the `#fragment`, defaults an empty path to `/`) so the same page isn't fetched twice under different spellings.
- Keeps a `visited` set and claims a URL the moment it's queued, so the same page is never fetched twice under different spellings.
- Runs the crawl through a `jobs` channel that a worker goroutine drains — the channel-based setup that a pool of workers will plug into next.
- Stops gracefully on Ctrl+C: the context is cancelled, in-flight requests are aborted, and the program exits cleanly.

## What it doesn't do yet

Being honest about the current state:

- **Link discovery is not wired up to the queue.** The workers fetch pages and count the links they find, but discovered links aren't enqueued yet, so in practice only the seed page gets fetched. This is the next milestone (see [Roadmap](#roadmap)).
- **No politeness controls.** There's no robots.txt handling, no per-host delay and no user-agent header, so please don't point it at a large site.
- **No persistent output.** Results are printed to stdout; nothing is written to disk and there's no way to resume a crawl later.
- **Only `href` attributes are parsed**, via a regular expression. Links inside `<script>`, inline styles or JSON blobs are ignored.

## Requirements

- Go 1.25 or newer (the code uses `sync.WaitGroup.Go`, added in Go 1.25). It was developed against Go 1.27.1.

## Build and run

Clone the repo and build the binary:

```bash
git clone https://github.com/vedantlavale/growler.git
cd growler
go build -o growler .
```

Then point it at a site:

```bash
./growler -seed https://example.com -max 10
```

Or run it straight from the source tree without building a binary:

```bash
go run . -seed https://example.com -max 10
```

## Command-line flags

| Flag | Default | What it does |
| --- | --- | --- |
| `-seed` | `https://vedantworks.com` | The URL the crawl starts from. |
| `-max` | `10` | The maximum number of pages to crawl. |

## Example run

```
$ go run . -seed https://example.com -max 3
https://example.com 3
crawling URL:  https://example.com/
links found:  0
Sequential crawl finished.
```

The first line echoes back the seed and the limit so you can confirm your flags were picked up. After that, each page prints as it is picked up by a worker, followed by how many links were found on it. When there is nothing left to crawl, the run ends.

The work happens in the background, so the "crawling URL" lines can appear in any order once more than one worker is running — a later page may well be finished before an earlier one.

Press Ctrl+C at any point to stop. The program cancels the context, aborts any requests that are still in flight, and returns without a stack trace.

## How it works

The crawler is split into two files.

`main.go` is the entry point. It parses the flags, builds a `Crawler`, and — importantly — creates the root context:

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
defer stop()
c.Crawl(ctx)
```

`signal.NotifyContext` gives you a context that is cancelled automatically when the process receives SIGINT, which on a terminal is what Ctrl+C sends. Every blocking operation downstream selects on `ctx.Done()`, so a single cancellation unwinds the whole crawl.

`crawler/crawler.go` holds the crawler itself:

- **`fetch`** performs one HTTP GET with the context attached, and reads the whole body into a string. The context matters here too: cancelling it aborts the request instead of waiting for a server that may never answer.
- **`extractLinks`** finds `href="..."` attributes with a regular expression and resolves each one against the page's URL using `url.URL.ResolveReference`. That single call is what turns `/about` into `https://example.com/about`. Links with a non-HTTP scheme or no host are dropped.
- **`normalizeURL`** canonicalizes a URL so `HTTPS://Example.com`, `https://example.com/` and `https://example.com#top` all collapse to the same key. Scheme and host are case-insensitive; the path is not, so `/Docs` and `/docs` stay distinct.
- **`markVisited`** is the check-and-set helper that guards the `visited` map. It's deliberately a single function so that the lock lives in exactly one place — once several workers are running, "have I seen this URL?" and "mark it as seen" must be one atomic step, otherwise two workers can grab the same URL at the same time.
- **`worker`** is the loop each goroutine runs: take a URL off the `jobs` channel, fetch it, extract its links. It exits when the channel is closed or the context is cancelled.
- **`Crawl`** starts the workers, hands them the seed URL, and waits for them to finish with a `sync.WaitGroup`.

The concurrency shape is a producer/consumer setup: `jobs` is a buffered channel of URLs, workers read from it, and the coordinator closes the channel once there's nothing left to hand out. Closing the channel is how the workers learn to stop — receiving from a closed channel returns the zero value, which the worker loop treats as "no more work".

Today a single worker goroutine drains the channel. The next milestone is to start several of them against the same channel, which is why the queue is a channel in the first place.

## Project layout

```
.
├── main.go            # Flags, signal handling, wiring
├── crawler/
│   └── crawler.go     # The crawler: fetch, parse, normalize, workers
├── go.mod
└── .walkthrough/      # Personal learning notes and the build roadmap
```

The `.walkthrough` directory is where the project keeps its plan: `roadmap.md` lists the milestones and what's done, `notes.md` records the technical decisions and the reasons behind them, and `gaps.md` lists the things that needed fixing along the way. It's the history of how the crawler got here, and it's the first place to look if you want to follow the project from the beginning.

## Roadmap

The plan is broken into milestones, tracked in `.walkthrough/roadmap.md`:

- **M1 — Skeleton and cancellation** — done
- **M2 — Fetch and parse** — done
- **M3 — Visited dedupe** — done
- **M4 — Jobs queue and worker pool** — in progress
- **M5 — Completion tracking** — not started
- **M6 — End-to-end run** — not started

The immediate next step is finishing M4: running several workers against the shared channel and having the producer enqueue the seed plus every link the workers discover.

## Contributing

This is a personal learning project, but issues and pull requests are welcome — especially if you can help squash a bug or explain a Go concept more clearly than the comments do.

If you'd like to pick something up, look at `.walkthrough/roadmap.md` for the open milestones, then open an issue describing what you want to work on before you start.

Before opening a pull request:

```bash
go build ./...
go vet ./...
gofmt -l .
```

`gofmt -l .` should print nothing.

## License

No license has been chosen yet.

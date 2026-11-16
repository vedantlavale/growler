package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/vedantlavale/growler/crawler"
)

func main() {

	seed := flag.String("seed", "https://vedantworks.com", "seed url") //flag is the go standard library used to parse command line flags

	//we define each flag name, default value, and help text

	maxPages := flag.Int("max", 10, "max pages to crawl")
	flag.Parse()

	fmt.Println(*seed, *maxPages)

	c := crawler.NewCrawler(*seed, *maxPages)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	// now it gives you a ctx that becomes canceled when SIGINT (Ctrl+C, os.Interrupt) happens. You defer stop() to clean up the signal handler, and you pass that same ctx into c.Crawl(ctx). The call looks the same, but the context you handed over can now carry a "stop" signal.

	defer stop()
	c.Crawl(ctx)
}

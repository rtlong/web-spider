package main

import (
	"flag"
	"fmt"
	"time"

	"net/url"
	"os"

	"github.com/rtlong/web-spider/spider"
)

var (
	concurrency = flag.Int("c", 100, "Max number of simultaneous open connections")
	jsonOutput  = flag.Bool("j", false, "Dump output as JSON to get much more information that the default summary output")
	depth       = flag.Int("d", 20, "Maximum depth of spidering (-1 indicates no limit)")
	redundancy  = flag.Int("r", 1, "Max number of fetches per URL")
	maxURLs     = flag.Int("m", 200000, "Max number of unique URLs to request")
	timeout     = flag.Duration("t", time.Second*20, "Timeout for any request")
	seedURL     url.URL
	logger      Logger
)

func main() {
	flag.Parse()

	if *jsonOutput {
		logger = new(JSONLogger)
	} else {
		logger = new(PlaintextLogger)
	}
	logger.SetOutput(os.Stdout)

	seedURL, err := url.Parse(flag.Arg(0))
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to parse input URL: %s", err))
	} else if !seedURL.IsAbs() {
		logger.Fatal("You must supply a URL to start with")
	}

	results := make(chan *spider.Result)

	s := spider.Spider{
		Fetcher: &spider.SimpleHTMLFetcher{
			Timeout: *timeout,
		},
		Results:     results,
		MaxDepth:    *depth,
		Concurrency: *concurrency,
		Redundancy:  *redundancy,
		MaxURLs:     *maxURLs,
		LinkFilterFunc: func(l spider.Link) bool {
			return l.URL.Host == seedURL.Host
		},
	}

	go func() {
		s.Crawl(seedURL)
		close(s.Results)
	}()

	logResults(results)
}

func logResults(results <-chan *spider.Result) {
	for r := range results {
		logger.PrintResult(r)
	}
}

package main

import (
	"flag"
	"fmt"

	"net/url"
	"os"

	"github.com/rtlong/web-spider/spider"
)

var (
	concurrency = flag.Int("c", 100, "Max number of simultaneous open connections")
	jsonOutput  = flag.Bool("j", false, "Dump output as JSON to get much more information that the default summary output")
	depth       = flag.Int("d", 20, "Maximum depth of spidering (-1 indicates no limit)")
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

	go logResults(results)

	s := spider.Spider{
		Fetcher:     spider.SimpleHTMLFetcher{},
		Results:     results,
		MaxDepth:    *depth,
		Concurrency: *concurrency,
		Redundancy:  1,
	}
	s.Crawl(seedURL)
	close(s.Results)
}

func logResults(results <-chan *spider.Result) {
	for r := range results {
		logger.PrintResult(r)
	}
}

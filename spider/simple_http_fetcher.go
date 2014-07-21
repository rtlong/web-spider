package spider

import (
	"io"
	"net/http"
	"net/url"

	"code.google.com/p/go.net/html"

	"github.com/PuerkitoBio/goquery"
)

type SimpleHTMLFetcher struct{}

// Fetch one Job, return results over Results channel
func (f SimpleHTMLFetcher) Fetch(job *Job) (result *Result) {
	result = &Result{Job: *job}

	req, err := http.NewRequest(job.Method, job.URL.String(), nil)
	if err != nil {
		result.Error = NewError("creating http.Request", err)
		return
	}
	req.Header.Add("User-Agent", "github.com/rtlong/web-spider")

	result.RecordStart()
	resp, err := client.Do(req)
	result.RecordEnd()
	if err != nil {
		result.Error = NewError("request", err)
		return
	}
	defer resp.Body.Close()

	result.Response = resp

	// OPTIMIZE: possibly use a channel here to queue goroutines as the page is
	// parsed?
	links, err := enumerateLinks(resp.Request.URL, resp.Body)
	if err != nil {
		result.Error = NewError("parsing", err)
		return
	}
	result.Links = links

	return
}

func enumerateLinks(contextURL *url.URL, r io.Reader) ([]*url.URL, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	aElements := doc.Find("a")

	urls := make([]*url.URL, 0, aElements.Length())

	for i := 0; i < aElements.Length(); i++ {
		node := aElements.Get(i)

		href := hrefAttrValue(node)
		if href == "" {
			continue
		}

		parsed, err := contextURL.Parse(href)
		if err != nil || !(parsed.Scheme == "http" || parsed.Scheme == "https") {
			continue
		}
		// Ignore the part of the URL after the "#"
		parsed.Fragment = ""

		urls = append(urls, parsed)
	}
	return urls, nil
}

func hrefAttrValue(n *html.Node) string {
	for _, attr := range n.Attr {
		if attr.Key == "href" {
			return attr.Val
		}
	}
	return ""
}

package spider

import (
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"code.google.com/p/go.net/html"

	"github.com/PuerkitoBio/goquery"
)

type Href string

type SimpleHTMLFetcher struct {
	Client      http.Client
	Timeout     time.Duration
	initialized bool
}

func (f *SimpleHTMLFetcher) init() {
	if f.initialized {
		return
	}

	// create an http.Client with the timeouts built in
	f.Client = http.Client{
		Transport: &http.Transport{
			Dial: func(n, a string) (net.Conn, error) {
				return net.DialTimeout(n, a, f.Timeout)
			},
			ResponseHeaderTimeout: f.Timeout,
		},
	}

	f.initialized = true
}

func (f *SimpleHTMLFetcher) Fetch(job *Job, links chan<- Link) (result *Result) {
	f.init()
	var err error

	result = &Result{Job: *job}

	req, err := http.NewRequest(job.Method, job.URL.String(), nil)
	if err != nil {
		result.Error = NewError("creating http.Request", err)
		return
	}
	req.Header.Add("User-Agent", "github.com/rtlong/web-spider")

	result.RecordStart()
	resp, err := f.Client.Do(req)
	result.RecordEnd()
	if err != nil {
		result.Error = NewError("request", err)
		return
	}
	defer resp.Body.Close()

	result.Response = resp

	if job.Depth != 0 && job.Method == MethodGET && responseIsHTML(resp) {
		hrefs := make(chan Href)
		go sendLinks(job, resp.Request.URL, hrefs, links)

		err = findHrefs(resp.Body, hrefs)
		close(hrefs)
		if err != nil {
			result.Error = NewError("parsing", err)
			return
		}
	} else {
		close(links)
	}

	return
}

func responseIsHTML(resp *http.Response) bool {
	ct := resp.Header.Get("Content-Type")
	if i := strings.Index(ct, ";"); i > 0 {
		ct = ct[:i]
	}
	return ct == "text/html"
}

func sendLinks(job *Job, ctxURL *url.URL, hrefs <-chan Href, links chan<- Link) {
	for href := range hrefs {
		parsed, err := ctxURL.Parse(string(href))
		if err != nil || !(parsed.Scheme == "http" || parsed.Scheme == "https") {
			continue
		}
		// Ignore the part of the URL after the "#"
		parsed.Fragment = ""

		links <- Link{URL: parsed, Job: job}
	}
	close(links)
}

func findHrefs(r io.Reader, hrefs chan<- Href) error {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return err
	}

	aElements := doc.Find("a")
	for i := 0; i < aElements.Length(); i++ {
		node := aElements.Get(i)
		href := hrefAttrValue(node)
		if href == "" {
			continue
		}
		hrefs <- Href(href)
	}

	return nil
}

func hrefAttrValue(n *html.Node) string {
	for _, attr := range n.Attr {
		if attr.Key == "href" {
			return attr.Val
		}
	}
	return ""
}

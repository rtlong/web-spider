package spider

import (
	"net/http"
	"net/url"
	"sync"
)

const (
	MethodGET  = "GET"
	MethodHEAD = "HEAD"
)

var (
	client = &http.Client{}
)

type Spider struct {
	Fetcher     Fetcher
	Results     chan *Result
	MaxDepth    int
	Redundancy  int
	Concurrency int
	sem         chan bool
	urlMap      map[string]int
	wg          sync.WaitGroup
	mx          sync.Mutex
	initialized bool
}

func (s *Spider) init() {
	if s.initialized {
		return
	}

	s.sem = make(chan bool, s.Concurrency)
	for i := 0; i < s.Concurrency; i++ {
		s.sem <- true
	}

	s.urlMap = make(map[string]int)
	s.wg = sync.WaitGroup{}
	s.mx = sync.Mutex{}

	s.initialized = true
	return
}

func (s *Spider) Crawl(startURL *url.URL) {
	s.init()

	s.wg.Add(1)
	go s._crawl(&Job{URL: *startURL, Method: MethodGET}, s.MaxDepth)

	s.wg.Wait()
}

func (s *Spider) _crawl(job *Job, depth int) {
	defer s.wg.Done()
	if count := s.urlMap[job.URL.String()]; count >= s.Redundancy {
		return
	}

	<-s.sem
	result := s.Fetcher.Fetch(job)
	s.sem <- true

	s.mx.Lock()
	// increment fetch count for this URL
	s.urlMap[job.URL.String()]++
	if depth != 0 {
		// iterate the discovered URLs and queue them for fetching, unless they're already fetched
		for _, link := range result.Links {
			if link != nil && job.URL.Host == link.Host {
				if count := s.urlMap[link.String()]; count < s.Redundancy {
					var method = MethodGET
					if count > 0 {
						method = MethodHEAD
					}
					s.wg.Add(1)
					go s._crawl(&Job{URL: *link, Method: method}, depth-1)
				}
			}
		}
	}
	s.mx.Unlock()

	s.Results <- result
}

type Fetcher interface {
	Fetch(*Job) *Result
}

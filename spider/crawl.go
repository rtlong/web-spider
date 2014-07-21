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

type Fetcher interface {
	Fetch(*Job) *Result
}

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

func (s *Spider) Crawl(startURL *url.URL) {
	s.init()

	s.wg.Add(1)
	go s._crawl(&Job{URL: *startURL, Method: MethodGET}, s.MaxDepth)

	s.wg.Wait()
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

func (s *Spider) _crawl(job *Job, depth int) {
	defer s.wg.Done()

	s.mx.Lock()
	// increment fetch count for this URL, before the attempt to fetch
	s.urlMap[job.URL.String()]++
	count := s.urlMap[job.URL.String()]
	s.mx.Unlock()

	if s.Redundancy > 0 && count > s.Redundancy {
		return
	}

	// Throttled by the semaphore, fetch
	<-s.sem
	result := s.Fetcher.Fetch(job)
	s.sem <- true

	if job.Method == MethodGET && depth != 0 {
		s.mx.Lock()
		for _, link := range result.Links {
			if job.URL.Host != link.Host {
				continue
			}
			count := s.urlMap[link.String()]
			if s.Redundancy > 0 && count >= s.Redundancy {
				continue
			}
			var method = MethodGET
			if count > 0 {
				method = MethodHEAD
			}
			s.wg.Add(1)
			go s._crawl(&Job{URL: *link, Method: method}, depth-1)
		}
		s.mx.Unlock()
	}

	s.Results <- result
}

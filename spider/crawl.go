package spider

import (
	"net/url"
	"sync"
)

const (
	MethodGET  = "GET"
	MethodHEAD = "HEAD"
)

type Fetcher interface {
	Fetch(*Job, chan<- Link) *Result
}

// A discovered outgoing link from a fetched HTML page
type Link struct {
	URL *url.URL
	Job *Job // Fetch job from which this link was found
}

type Spider struct {
	Fetcher        Fetcher
	Results        chan *Result
	MaxDepth       int
	Redundancy     int
	Concurrency    int
	MaxURLs        int
	LinkFilterFunc func(Link) bool
	sem            chan bool
	urlMap         map[string]int
	wg             sync.WaitGroup
	mx             sync.Mutex
	initialized    bool
}

func (s *Spider) Crawl(startURL *url.URL) {
	s.init()

	s.wg.Add(1)
	go s.fetch(&Job{URL: *startURL, Method: MethodGET, Depth: s.MaxDepth})

	s.wg.Wait()
}

func (s *Spider) enqueue(links chan Link, done chan bool) {
	for link := range links {
		if len(s.urlMap) >= s.MaxURLs {
			continue
		}

		if s.LinkFilterFunc != nil && !s.LinkFilterFunc(link) {
			continue
		}

		s.mx.Lock()
		count := s.urlMap[link.URL.String()]
		s.mx.Unlock()

		if s.Redundancy > 0 && count >= s.Redundancy {
			continue
		}

		var method = MethodGET
		if count > 0 {
			method = MethodHEAD
		}

		s.wg.Add(1)
		go s.fetch(&Job{URL: *link.URL, Method: method, Depth: link.Job.Depth - 1})
	}
	done <- true
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
}

func (s *Spider) fetch(job *Job) {
	defer s.wg.Done()

	s.mx.Lock()
	// increment fetch count for this URL, before the attempt to fetch
	s.urlMap[job.URL.String()]++
	count := s.urlMap[job.URL.String()]
	s.mx.Unlock()

	if s.Redundancy > 0 && count > s.Redundancy {
		return
	}

	links := make(chan Link)
	enqueueDone := make(chan bool)
	go s.enqueue(links, enqueueDone)

	// Throttled by the semaphore, fetch
	<-s.sem
	result := s.Fetcher.Fetch(job, links)
	s.sem <- true

	s.Results <- result

	// Wait until enqueue is done
	<-enqueueDone
}

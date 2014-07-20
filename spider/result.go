package spider

import (
	"net/http"
	"net/url"
	"time"
)

type Result struct {
	Job             Job
	Response        *http.Response
	Time            time.Time
	RequestDuration time.Duration
	Error           Error
	Links           []*url.URL
}

func (r *Result) RecordStart() {
	r.Time = time.Now()
}

func (r *Result) RecordEnd() {
	r.RequestDuration = time.Since(r.Time)
}

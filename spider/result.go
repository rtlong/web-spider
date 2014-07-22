package spider

import (
	"net/http"
	"time"
)

type Result struct {
	Job      Job
	Response *http.Response
	Time     time.Time
	Duration time.Duration
	Error    Error
}

func (r *Result) RecordStart() {
	r.Time = time.Now()
}

func (r *Result) RecordEnd() {
	r.Duration = time.Since(r.Time)
}

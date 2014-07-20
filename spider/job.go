package spider

import (
	"fmt"
	"net/url"
)

type Job struct {
	URL    url.URL
	Method string
}

func (j Job) String() string {
	return fmt.Sprintf("%s %s", j.Method, j.URL.String())
}

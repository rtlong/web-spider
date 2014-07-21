package spider

import (
	"fmt"
	"net/url"
)

type Job struct {
	URL    url.URL
	Method string
	Depth  int
}

func (j Job) String() string {
	return fmt.Sprintf("%s %s", j.Method, j.URL.String())
}

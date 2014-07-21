# web-spider

This is a web spider. It's very much a work in progress, but it's working well
enough to perhaps be useful to someone. I'm new to Go, so please excuse any
glaring/horrid mistakes.

## Goals

The intention of this project was to create a spider much like a search
engine's, with the exception that I'm not interested in saving or indexing the
fetched pages. This spider is meant for scanning a site, verifying that there
are no broken links, no dead pages, and collecting response time and other
stats about the response, but not necessarily saving the response itself.

## Usage

```shell
% go get github.com/rtlong/web-spider
% web-spider http://example.com
```

## TODO

- check `<link>`, `<img>`, `<script>`, and `<iframe>` tag hrefs in addition to `<a>`
- ensure `href="//blah.com/foo"` urls are not ignored due to `URL.Scheme` assertion
- add tests!
- improve output
- add more configurability:
    - ability to add extra headers during requests

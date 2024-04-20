package main

import (
	"net/url"
	"path/filepath"
	"strings"
)

const WIKI = "https://en.wikipedia.org/wiki/"

type Pages []string

// TODO: Filter namespace
func getPages(links []string) Pages {
	pages := make([]string, 0)
	for _, to := range links {
		if strings.HasPrefix(to, WIKI) {
			toPage, _ := filepath.Rel(WIKI, to)
			if !strings.ContainsAny(toPage, ":#") {
				res, err := url.QueryUnescape(toPage)
				if err != nil {
					res = toPage
				}
				pages = append(pages, res)
			}
		}
	}

	return pages
}

func getLinks(page string) Pages {
	return getPages(scrap(WIKI + page))
}

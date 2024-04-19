package main

import (
	"net/url"
	"path/filepath"
	"strings"
)

const WIKI = "https://en.wikipedia.org/wiki/"

func getPages(links []string) []string {
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

func getLinks2(pages []string, channel chan Links) {
	for _, page := range pages {
		url := WIKI + page

		channel <- Links{
			From: page,
			To:   getPages(scrap(url)),
		}
	}
}

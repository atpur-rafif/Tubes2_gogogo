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
	visited := make(map[string]bool)

	for _, to := range links {
		if strings.HasPrefix(to, WIKI) {
			toPage, _ := filepath.Rel(WIKI, to)
			if !strings.ContainsAny(toPage, ":#") {
				page, err := url.QueryUnescape(toPage)
				if err != nil {
					page = toPage
				}

				if visited[page] {
					continue
				}
				visited[page] = true
				pages = append(pages, page)
			}
		}
	}

	return pages
}

// TODO: Redirect map
func getLinks(page string) Pages {
	// P := make(map[string][]string)
	// P["Hitler"] = []string{"B"}
	// P["B"] = []string{"C"}
	// P["C"] = []string{"D"}
	// P["D"] = []string{"E"}
	// P["E"] = []string{"Traffic"}
	// return P[page]
	return getPages(scrap(WIKI + page))
}

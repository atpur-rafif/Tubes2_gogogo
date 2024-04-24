package main

import (
	"net/url"
	"path/filepath"
	"strings"
	"sync"
)

const WIKI = "https://en.wikipedia.org/wiki/"

type Pages []string

var canonLinks = make(map[string]string)
var canonLinksMutex sync.Mutex

func getCanon(page string) string {
	var canon string
	canonLinksMutex.Lock()
	canon = canonLinks[page]
	canonLinksMutex.Unlock()
	return canon
}

func setCanon(page, canon string) {
	canonLinksMutex.Lock()
	canonLinks[page] = canon
	canonLinksMutex.Unlock()
}

func parsePage(to string) (string, bool) {
	if !strings.HasPrefix(to, WIKI) {
		return "", false
	}

	rel, err := filepath.Rel(WIKI, to)
	if err != nil {
		return "", false
	}

	if strings.ContainsAny(rel, ":#") {
		return "", false
	}

	page, err := url.QueryUnescape(rel)
	if err != nil {
		return "", false
	}

	return page, true
}

// TODO: Filter namespace
func filterPages(links []string) Pages {
	pages := make([]string, 0)

	visited := make(map[string]bool)
	for _, to := range links {
		page, ok := parsePage(to)

		if !ok {
			continue
		}

		if visited[page] {
			continue
		}
		visited[page] = true

		pages = append(pages, page)
	}

	return pages
}

func getLinks(page string) Pages {
	// P := make(map[string][]string)
	// P["Hitler"] = []string{"B"}
	// P["B"] = []string{"C"}
	// P["C"] = []string{"D"}
	// P["D"] = []string{"E"}
	// P["E"] = []string{"Traffic"}
	// return P[page]

	canonURL, pages := scrap(WIKI + page)
	canonPage, ok := parsePage(canonURL)
	if !ok {
		canonPage = page
	}
	setCanon(page, canonPage)

	return filterPages(pages)
}

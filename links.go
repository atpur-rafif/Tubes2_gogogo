package main

import (
	"net/url"
	"path/filepath"
	"strings"
)

const WIKI = "https://en.wikipedia.org/wiki/"

type Pages []string

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

func getLinks(page string) (string, Pages) {
	P := make(map[string][]string)
	P["Adolf_Hitler"] = []string{"B_"}
	P["Hitler"] = []string{"B_"}
	P["B"] = []string{"C", "Hitler"}
	P["B_"] = []string{"C", "Hitler"}
	P["C"] = []string{"D", "B_"}
	P["D"] = []string{"E", "B"}
	P["E"] = []string{"Traffic"}

	canon := page
	if page == "Hitler" {
		canon = "Adolf_Hitler"
	}
	if page == "Traffic_" {
		canon = "Traffic"
	}
	if page == "B_" {
		canon = "B"
	}

	return canon, P[page]

	// canonURL, pages := scrap(WIKI + page)
	// canonPage, ok := parsePage(canonURL)
	// if !ok {
	// 	canonPage = page
	// }
	//
	// return canonPage, filterPages(pages)
}

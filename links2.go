package main

import (
	"log"
	"net/url"
	"path/filepath"
	"strings"
)

func getLinks2(pages []string, channel chan Links) {
	for _, page := range pages {
		url, err := url.JoinPath(WIKI, page)
		if err != nil {
			log.Println("Can't join path " + url + " with " + page)
			return
		}

		resultPages := make([]string, 0)
		for _, to := range scrap(url) {
			if strings.HasPrefix(to, WIKI) {
				toPage, _ := filepath.Rel(WIKI, to)
				if !strings.ContainsAny(toPage, ":#") {
					resultPages = append(resultPages, toPage)
				}
			}
		}

		channel <- Links{
			From: page,
			To:   resultPages,
		}
	}
}

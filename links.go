package main

import (
	"crypto/sha256"
	"encoding/base32"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const WIKI = "https://en.wikipedia.org/wiki/"
const CACHE_DIR = "cache"

type Pages []string

func parsePage(to string) (string, bool) {
	if !strings.HasPrefix(to, WIKI) {
		return "", false
	}

	rel, err := filepath.Rel(WIKI, to)
	if err != nil {
		return "", false
	}

	if strings.Contains(rel, ":") {
		return "", false
	}

	// if strings.HasPrefix(rel, "Talk:") ||
	// 	// Namespace
	// 	strings.HasPrefix(rel, "User:") ||
	// 	strings.HasPrefix(rel, "Wikipedia:") ||
	// 	strings.HasPrefix(rel, "File:") ||
	// 	strings.HasPrefix(rel, "MediaWiki:") ||
	// 	strings.HasPrefix(rel, "Template:") ||
	// 	strings.HasPrefix(rel, "Help:") ||
	// 	strings.HasPrefix(rel, "Category:") ||
	// 	strings.HasPrefix(rel, "Portal:") ||
	// 	strings.HasPrefix(rel, "Draft:") ||
	// 	strings.HasPrefix(rel, "TimedText:") ||
	// 	strings.HasPrefix(rel, "Module:") ||
	// 	// Virtual namespace
	// 	strings.HasPrefix(rel, "Special:") ||
	// 	strings.HasPrefix(rel, "Media:") ||
	// 	// Former namespace
	// 	strings.HasPrefix(rel, "Book:") ||
	// 	strings.HasPrefix(rel, "Course:") ||
	// 	strings.HasPrefix(rel, "Institution:") ||
	// 	strings.HasPrefix(rel, "Topic:") {
	// 	return "", false
	// }

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

func hash(str string) string {
	hasher := sha256.New()
	hasher.Write([]byte(str))
	return base32.StdEncoding.EncodeToString(hasher.Sum(nil))
}

func fileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func createCachePath(str string) string {
	return CACHE_DIR + "/" + hash(str)
}

func readCache(page string) (string, []string, bool) {
	time.Sleep(100 * time.Millisecond)

	pagePath := createCachePath(page)
	if !fileExist(pagePath) {
		return "", nil, false
	}

	cacheByte, err := os.ReadFile(pagePath)
	cache := string(cacheByte)
	if err != nil {
		return "", nil, false
	}

	if len(cache) == 0 {
		os.Remove(pagePath)
		return "", nil, false
	}

	if cache[0] != '\n' {
		return page, strings.Split(cache, "\n"), true
	}

	canon := strings.Replace(cache, "\n", "", 1)
	canonPath := createCachePath(canon)
	if !fileExist(canonPath) {
		return "", nil, false
	}

	cacheCanonByte, err := os.ReadFile(canonPath)
	if err != nil {
		return "", nil, false
	}

	cacheCanon := string(cacheCanonByte)
	return canon, strings.Split(cacheCanon, "\n"), true
}

func writeCache(page, canon string, pages []string) {
	os.MkdirAll("cache", os.ModePerm)
	go func() {
		pagePath := createCachePath(page)
		canonPath := createCachePath(canon)

		if page != canon {
			if fileExist(pagePath) {
				return
			}
			os.WriteFile(pagePath, []byte("\n"+canon), os.ModePerm)
		}

		if fileExist(canonPath) {
			return
		}
		os.WriteFile(canonPath, []byte(strings.Join(pages, "\n")), os.ModePerm)
	}()
}

func getLinks(page string) (string, Pages) {
	var canon string
	var pages []string

	// canon, pages, ok := readCache(page)
	// if ok {
	// 	return canon, pages
	// }

	canonURL, pagesURL := scrap(WIKI + url.PathEscape(page))
	pages = filterPages(pagesURL)
	canon, parseOk := parsePage(canonURL)
	if !parseOk {
		canon = page
	}

	// writeCache(page, canon, pages)

	return canon, pages
}

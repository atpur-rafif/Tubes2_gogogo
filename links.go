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
	pagePath := createCachePath(page)
	if !fileExist(pagePath) {
		return "", nil, false
	}

	time.Sleep(100 * time.Millisecond)

	cacheByte, _ := os.ReadFile(pagePath)
	cache := string(cacheByte)

	if cache[0] != '\n' {
		return page, strings.Split(cache, "\n"), true
	}

	canon := strings.Replace(cache, "\n", "", 1)
	canonPath := createCachePath(canon)
	cacheCanonByte, _ := os.ReadFile(canonPath)
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
	// time.Sleep(300 * time.Millisecond)
	// P := make(map[string][]string)
	// P["Adolf_Hitler"] = []string{"B_"}
	// P["Hitler"] = []string{"B_"}
	// P["B"] = []string{"C", "Hitler", "D"}
	// P["B_"] = []string{"C", "Hitler", "D"}
	// P["C"] = []string{"D", "B_"}
	// P["D"] = []string{"E", "B", "F"}
	// P["E"] = []string{"Traffic"}
	// P["F"] = []string{"Traffic"}
	//
	// canon := page
	// if page == "Hitler" {
	// 	canon = "Adolf_Hitler"
	// }
	// if page == "Traffic_" {
	// 	canon = "Traffic"
	// }
	// if page == "B_" {
	// 	canon = "B"
	// }
	//
	// return canon, P[page]

	canon, pages, ok := readCache(page)
	if ok {
		return canon, pages
	}

	canonURL, pageURL := scrap(WIKI + page)
	pages = filterPages(pageURL)
	canon, ok = parsePage(canonURL)
	if !ok {
		canon = page
	}

	writeCache(page, canon, pages)
	return canon, pages
}

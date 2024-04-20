package main

import (
	"bytes"
	"log"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func toAbsUrl(from, to *url.URL) url.URL {
	toStr := to.String()
	if to.IsAbs() {
		return *to
	} else {
		to.Scheme = from.Scheme
		if !strings.HasPrefix(toStr, "//") {
			to.Host = from.Host
		}
		return *to
	}
}

func scrap(urlStr string) []string {
	result := make([]string, 0)
	from, err := url.Parse(urlStr)
	if err != nil {
		log.Println("Can't parse URL " + urlStr)
		return result
	}

	log.Println("[Scrapper] Visiting  " + urlStr)
	response, err := http.Get(urlStr)
	if err != nil {
		log.Println("Can't visit URL " + urlStr)
		return result
	}
	defer response.Body.Close()

	insideMain := false
	tokenizer := html.NewTokenizer(response.Body)
	count := 0
	for {
		token := tokenizer.Next()
		if token == html.ErrorToken {
			break
		}

		name, _ := tokenizer.TagName()
		if bytes.Equal(name, []byte("main")) {
			if token == html.StartTagToken {
				insideMain = true
			} else if token == html.EndTagToken {
				insideMain = false
			}
		}

		if !insideMain {
			continue
		}

		if bytes.Equal(name, []byte("a")) {
			for {
				key, value, next := tokenizer.TagAttr()
				if bytes.Equal(key, []byte("href")) {
					str := string(value)
					to, err := url.Parse(str)
					if err != nil {
						log.Println("Can't parse URL " + str)
						continue
					}
					absTo := toAbsUrl(from, to)
					result = append(result, absTo.String())
				}

				if !next {
					break
				}
			}
			count += 1
		}
	}
	log.Println("[Scrapper] Links found in "+urlStr+":", len(result))

	return result
}

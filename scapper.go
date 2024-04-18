package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func toAbsUrl(from, to *url.URL) string {
	toStr := to.String()
	if to.IsAbs() {
		return toStr
	} else {
		if strings.HasPrefix(toStr, "//") {
			return from.Scheme + ":" + toStr
		} else {
			return from.Scheme + "://" + from.Host + toStr
		}
	}
}

func scrap() {
	urlStr := "https://en.wikipedia.org/wiki/Main_Page"
	from, err := url.Parse(urlStr)
	if err != nil {
		log.Println("Can't parse URL " + urlStr)
		return
	}

	response, err := http.Get(urlStr)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	tokenizer := html.NewTokenizer(response.Body)
	for {
		token := tokenizer.Next()
		if token == html.ErrorToken {
			break
		}

		name, _ := tokenizer.TagName()
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
					absUrl := toAbsUrl(from, to)
					fmt.Println(absUrl)
				}

				if !next {
					break
				}
			}
		}
	}
}

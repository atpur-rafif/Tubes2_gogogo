package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const API = "https://en.wikipedia.org/w/api.php"

type WikipediaResponse struct {
	Continue *(struct {
		Continue   string
		Plcontinue string
	})
	Query struct {
		Pages map[string](struct {
			Title string
			Links [](struct {
				Title string
			})
		})
	}
}

func getLinks(pages []string) map[string][]string {
	query := url.Values{}
	query.Add("action", "query")
	query.Add("format", "json")
	query.Add("prop", "links")
	query.Add("pllimit", "max")
	query.Add("titles", strings.Join(pages, "|"))

	links := make(map[string][]string)
	for {
		url := fmt.Sprintf("%s?%s", API, query.Encode())
		response, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		defer response.Body.Close()
		byte, err := io.ReadAll(response.Body)
		if err != nil {
			panic(err)
		}

		var parsed WikipediaResponse
		err = json.Unmarshal(byte, &parsed)
		if err != nil {
			fmt.Println(err)
		}

		for _, page := range parsed.Query.Pages {
			for _, link := range page.Links {
				links[page.Title] = append(links[page.Title], link.Title)
			}
		}

		if parsed.Continue == nil {
			break
		} else {
			query.Set("plcontinue", parsed.Continue.Plcontinue)
			query.Set("continue", parsed.Continue.Continue)
		}
	}

	return links
}

func main() {
	fmt.Println("Hello, world!")
	links := getLinks([]string{"Kirby|Short, Mississippi"})
	fmt.Println(links)
}

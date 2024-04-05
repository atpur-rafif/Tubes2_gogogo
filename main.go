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
		}
	}

	return links
}

func main() {
	start := "Wikipedia"
	end := "Knowledge"

	found := false
	var foundPath []string
	fmt.Println("Started")

	traversed := make(map[string]bool)
	stack := make([][]string, 0)
	stack = append(stack, []string{start})
	for {
		batch := make([][]string, 0)
		tails := make([]string, 0)
		for len(stack) > 0 && len(batch) < 50 {
			path := stack[0]
			stack = stack[1:]

			top := path[len(path)-1]
			if traversed[top] {
				continue
			}
			traversed[top] = true

			batch = append(batch, path)
			tails = append(tails, top)
		}

		links := getLinks(tails)
		for _, path := range batch {
			top := path[len(path)-1]
			for _, page := range links[top] {
				newPath := make([]string, len(path))
				copy(newPath, path)
				newPath = append(newPath, page)
				stack = append(stack, newPath)
				fmt.Println(newPath)

				if page == end {
					found = true
					foundPath = newPath
				}
			}

			if found {
				break
			}
		}
		if found {
			break
		}
	}
	fmt.Println(foundPath)
	fmt.Println("Finished")
}

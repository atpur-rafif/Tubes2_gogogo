package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Links struct {
	From string
	To   []string
}

// TODO: Cancelation
func getLinks(pages []string, channel chan Links) {
	query := url.Values{}
	query.Add("action", "query")
	query.Add("format", "json")
	query.Add("prop", "links")
	query.Add("pllimit", "max")
	query.Add("titles", strings.Join(pages, "|"))

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
			to := make([]string, 0)
			for _, link := range page.Links {
				to = append(to, link.Title)
			}

			channel <- Links{
				From: page.Title,
				To:   to,
			}
		}

		if parsed.Continue == nil {
			break
		} else {
			query.Set("plcontinue", parsed.Continue.Plcontinue)
		}
	}

	close(channel)
}

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const API = "https://en.wikipedia.org/w/api.php"

type WikipediaResponse struct {
	Continue struct {
		Continue   string
		Plcontinue string
	} `json:"continue"`
	Query struct {
		Pages map[string](struct {
			Title string
			Links [](struct {
				Title string
			})
		})
	}
}

func getLinks(page string) []string {
	query := url.Values{}
	query.Add("action", "query")
	query.Add("format", "json")
	query.Add("prop", "links")
	query.Add("pllimit", "max")
	query.Add("titles", page)

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

	fmt.Println(parsed)
	fmt.Println(url)

	return make([]string, 0)
}

func main() {
	fmt.Println("Hello, world!")
	getLinks("TypeScript")
}

package main

import "encoding/json"

const API = "https://en.wikipedia.org/w/api.php"
const WIKI = "https://en.wikipedia.org/wiki"

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

const (
	Update = iota + 1
	Started
	Finished
	Error
)

type Status uint8

type Request struct {
	Start string
	End   string
}

type Response struct {
	Status  Status `json:"status"`
	Message string `json:"message"`
}

func (s Status) String() string {
	switch s {
	case Update:
		return "update"
	case Started:
		return "started"
	case Finished:
		return "finished"
	default:
		return "error"
	}
}

func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

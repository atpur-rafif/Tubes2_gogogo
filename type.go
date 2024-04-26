package main

import "encoding/json"

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

const (
	Log = iota + 1
	Start
	Found
	End
	Error
)

type Status uint8

type Request struct {
	Start string
	End   string
	Force bool
	Type  string
}

type Response struct {
	Status  Status      `json:"status"`
	Message interface{} `json:"message"`
}

func (s Status) String() string {
	switch s {
	case Log:
		return "update"
	case Start:
		return "started"
	case Found:
		return "found"
	case End:
		return "finished"
	default:
		return "error"
	}
}

func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

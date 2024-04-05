package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const (
	Update = iota + 1
	Started
	Finished
	Error
)

type Status uint8

type ClientRequest struct {
	Start string
	End   string
}

type ClientResponse struct {
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

func main() {
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				break
			}

			if msgType == websocket.TextMessage {
				var request ClientRequest
				json.Unmarshal(msg, &request)

				if len(request.Start) == 0 || len(request.End) == 0 {
					conn.WriteJSON(&ClientResponse{
						Status:  Error,
						Message: `Empty "start" or "end" of field`,
					})
				}

				log.Println(request)
			}

		}
	})

	http.Handle("/", http.FileServer(http.Dir("./static")))
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		panic(err)
	}
}

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

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

func run(start, end string, channel chan string) {
	time.Sleep(5 * time.Second)
	channel <- start + ";" + end
	close(channel)
}

func main() {
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		write := make(chan Response)
		read := make(chan Request)
		quit := make(chan bool)

		go func(conn *websocket.Conn, read chan Request, write chan Response, quit chan bool) {
			for {
				msgType, msg, err := conn.ReadMessage()
				if err != nil {
					log.Println(err)
					quit <- true
					break
				}

				if msgType == websocket.TextMessage {
					var request Request
					json.Unmarshal(msg, &request)

					if len(request.Start) == 0 || len(request.End) == 0 {
						write <- Response{
							Status:  Error,
							Message: `Empty "start" or "end" of field`,
						}
						continue
					}

					read <- request
				}
			}
		}(conn, read, write, quit)

		go func(conn *websocket.Conn, read chan Request, write chan Response, quit chan bool) {
			for {
				select {
				case <-quit:
					break
				case msg := <-write:
					conn.WriteJSON(msg)
				}
			}
		}(conn, read, write, quit)

		for msg := range read {
			log.Println(msg)
			write <- Response{
				Status:  Started,
				Message: msg.Start + ";" + msg.End,
			}
		}
	})

	http.Handle("/", http.FileServer(http.Dir("./static")))
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		panic(err)
	}
}

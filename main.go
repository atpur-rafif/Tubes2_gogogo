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

func run(start, end string, channel chan Response, forceQuit chan bool) {
	linksChan := make(chan Links)

	goroutineEnded := make(chan bool)

	go func() {
		getLinks([]string{start, end}, linksChan)
		goroutineEnded <- true
	}()

L:
	for {
		select {
		case <-goroutineEnded:
			break L
		case <-forceQuit:
			break L
		case links := <-linksChan:
			from := links.from
			for _, to := range links.to {
				channel <- Response{
					Status:  Update,
					Message: from + " -> " + to,
				}
			}
		}
	}
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
		forceQuit := make(chan bool)

		go func(conn *websocket.Conn) {
			for {
				msgType, msg, err := conn.ReadMessage()
				if err != nil {
					log.Println(err)
					forceQuit <- true
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
		}(conn)

		go func(conn *websocket.Conn) {
			for {
				select {
				case <-forceQuit:
					break
				case msg := <-write:
					conn.WriteJSON(msg)
				}
			}
		}(conn)

		running := false
		finished := make(chan bool)
		for {
			select {
			case <-finished:
				running = false
			case <-forceQuit:
				break
			case req := <-read:
				if running {
					write <- Response{
						Status:  Error,
						Message: "Program still running",
					}
					continue
				}

				running = true
				go func() {
					run(req.Start, req.End, write, forceQuit)
					finished <- true
				}()
			}
		}
	})

	http.Handle("/", http.FileServer(http.Dir("./static")))
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		panic(err)
	}
}

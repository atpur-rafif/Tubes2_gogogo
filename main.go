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

func run(start, end string, channel chan Response, stop chan bool, _ chan bool) {
	channel <- Response{
		Status:  Started,
		Message: start + ";" + end,
	}

	for i := 0; i < 100; i++ {
		time.Sleep(100 * time.Millisecond)
		channel <- Response{
			Status:  Update,
			Message: "Update...",
		}
	}

	channel <- Response{
		Status:  Finished,
		Message: start + ";" + end,
	}
	stop <- true
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
		end := make(chan bool)
		stop := make(chan bool)

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
		}(conn, read, write, end)

		go func(conn *websocket.Conn, read chan Request, write chan Response, quit chan bool) {
			for {
				select {
				case <-quit:
					break
				case msg := <-write:
					conn.WriteJSON(msg)
				}
			}
		}(conn, read, write, end)

		running := false
		for {
			select {
			case <-end:
				break
			case <-stop:
				running = false
			case req := <-read:
				if running {
					write <- Response{
						Status:  Error,
						Message: "Program still running",
					}
					continue
				}

				go run(req.Start, req.End, write, stop, end)
				running = true
			}
		}
	})

	http.Handle("/", http.FileServer(http.Dir("./static")))
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		panic(err)
	}
}

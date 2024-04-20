package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type TraverseFunction func(string, string, chan Response, chan bool)

func run(start, end string, channel chan Response, forceQuit chan bool) {
	var fn TraverseFunction
	fn = SearchBFS
	fn(start, end, channel, forceQuit)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// func main() {
// 	forceQuit := make(chan bool)
// 	responses := make(chan Response)
// 	go SearchBFS("Highway", "Traffic", responses, forceQuit)
// 	for response := range responses {
// 		log.Println(response.Message)
// 	}
// }

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
	log.Println("Listening on port 3000")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		panic(err)
	}
}

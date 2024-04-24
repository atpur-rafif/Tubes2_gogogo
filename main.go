package main

import (
	"embed"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type TraverseFunction func(string, string, chan Response, chan bool)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//go:embed static/*
var static embed.FS

// func main() {
// 	forceQuit := make(chan bool)
// 	responses := make(chan Response)
// 	go func() {
// 		SearchBFS("Adolf_Hitler", "Traffic", responses, forceQuit)
// 		// SearchIDS("Highway", "Traffic", responses, forceQuit)
// 	}()
// 	for res := range responses {
// 		// _ = res
// 		log.Println(res)
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

					if request.Type != "BFS" && request.Type != "IDS" {
						write <- Response{
							Status:  Error,
							Message: "Invalid method",
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
		forceQuitRun := make(chan bool)
		for {
			select {
			case <-finished:
				running = false
			case <-forceQuit:
				log.Println("End")
				forceQuitRun <- true
				break
			case req := <-read:
				if running {
					if req.Force {
						forceQuitRun <- true
					} else {
						write <- Response{
							Status:  Error,
							Message: "Program still running",
						}
						continue
					}
				}

				running = true
				go func() {
					var fn TraverseFunction
					if req.Type == "BFS" {
						fn = SearchBFS
					} else if req.Type == "IDS" {
						fn = SearchIDS
					} else {
						log.Panic("Invalid method")
					}
					fn(req.Start, req.End, write, forceQuitRun)
					finished <- true
				}()
			}
		}
	})

	// content, _ := fs.Sub(static, "static")
	http.Handle("/", http.FileServer(http.Dir("static")))
	log.Println("Listening on port 3000")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		panic(err)
	}
}

package main

import (
	"embed"
	"encoding/json"
	// "io/fs"
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

func main_old() {
	// forceQuit := make(chan bool)
	// responses := make(chan Response)
	// go func() {
	// 	SearchBFS("Highway", "Traffic", responses, forceQuit)
	// 	// SearchIDS("Hitler", "Traffic", responses, forceQuit)
	// 	// SearchIDS("Highway", "Traffic", responses, forceQuit)
	// }()
	// for res := range responses {
	// 	_ = res
	// 	// log.Println(res)
	// }
}

func check200Status(page string) bool {
	req, err := http.Get(WIKI + page)
	if err != nil {
		return false
	}
	defer req.Body.Close()
	return req.StatusCode == 200
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
		wsQuit := make(chan bool)

		go func(conn *websocket.Conn) {
			for {
				msgType, msg, err := conn.ReadMessage()
				if err != nil {
					log.Println(err)
					wsQuit <- true
					break
				}

				if msgType == websocket.TextMessage {
					var request Request
					json.Unmarshal(msg, &request)

					if request.Cancel {
						read <- request
						continue
					}

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

					if !(check200Status(request.Start) && check200Status(request.End)) {
						write <- Response{
							Status:  Error,
							Message: "Page not found",
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
				case <-wsQuit:
					break
				case msg := <-write:
					conn.WriteJSON(msg)
				}
			}
		}(conn)

		running := false
		finished := make(chan bool)
		runQuit := make(chan bool)
		for {
			select {
			case <-finished:
				running = false
			case <-wsQuit:
				log.Println("End")
				runQuit <- true
				break
			case req := <-read:
				if running {
					if req.Cancel {
						runQuit <- true
					} else {
						write <- Response{
							Status:  Error,
							Message: "Program still running",
						}
					}
					continue
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
					fn(req.Start, req.End, write, runQuit)
					finished <- true
				}()
			}
		}
	})

	// content, _ := fs.Sub(static, "static")
	http.Handle("/", http.FileServer(http.Dir("static")))
	// http.Handle("/", http.FileServer(http.FS(content)))
	log.Println("Listening on port 3000")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		panic(err)
	}
}

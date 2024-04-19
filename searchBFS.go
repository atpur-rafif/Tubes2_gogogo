package main

import (
	"log"
	"strings"
)

func SearchBFS(start, end string, channel chan Response, forceQuit chan bool) {
	stack := make([][]string, 0)
	visited := make(map[string]bool)
	stack = append(stack, []string{start})

	var found []string
LO:
	for {
		if len(stack) == 0 {
			log.Println("Link from " + start + " to " + end + " not found")
			break
		}
		path := stack[0]
		stack = stack[1:]

		top := path[len(path)-1]
		if visited[top] {
			continue
		}
		visited[top] = true

		linksChan := make(chan Links)
		finished := make(chan bool)
		go func() {
			getLinks2([]string{top}, linksChan)
			finished <- true
		}()

	L:
		for {
			select {
			case <-finished:
				break L
			case <-forceQuit:
				break LO
			case link := <-linksChan:
				for _, to := range link.To {
					newPath := make([]string, len(path))
					copy(newPath, path)
					newPath = append(newPath, to)
					stack = append(stack, newPath)

					if to == end {
						found = newPath
						break LO
					}

					// channel <- Response{
					// 	Status:  Update,
					// 	Message: link.From + " ➡️ " + to,
					// }
				}

				channel <- Response{
					Status:  Update,
					Message: "Visited " + link.From,
				}
			}
		}
	}

	channel <- Response{
		Status:  Finished,
		Message: strings.Join(found, " ➡️ "),
	}
}

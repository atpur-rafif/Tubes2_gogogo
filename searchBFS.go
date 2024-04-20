package main

import (
	"log"
	"strings"
	"time"
)

const MAX_CONCURRENT = 10

type StateBFS struct {
	Stack   [][]string
	Visited map[string]bool
	Running int
}

type Visit struct {
	Path []string
	Next []string
}

func runStack(state *StateBFS, linksChan chan Visit) {
	for len((*state).Stack) > 0 && (*state).Running < MAX_CONCURRENT {
		path := (*state).Stack[0]
		(*state).Stack = (*state).Stack[1:]
		top := path[len(path)-1]

		if (*state).Visited[top] {
			continue
		}
		(*state).Visited[top] = true

		(*state).Running += 1
		go func() {
			linksChan <- Visit{
				Path: path,
				Next: getLinks2(top),
			}
		}()
	}
}

func SearchBFS(start, end string, channel chan Response, forceQuit chan bool) {
	var resultPath []string
	state := StateBFS{
		Stack:   make([][]string, 0),
		Visited: make(map[string]bool),
		Running: 0,
	}
	state.Stack = append(state.Stack, []string{start})

	visitChan := make(chan Visit)
	log.Println(state)
	runStack(&state, visitChan)

L:
	for {
		select {
		case <-forceQuit:
			break L
		case visit := <-visitChan:
			for _, next := range visit.Next {
				newPath := make([]string, len(visit.Path))
				copy(newPath, visit.Path)
				newPath = append(newPath, next)
				state.Stack = append(state.Stack, newPath)

				if next == end {
					resultPath = newPath
					break L
				}
			}

			channel <- Response{
				Status:  Update,
				Message: "Visited " + visit.Path[len(visit.Path)-1],
			}

			state.Running -= 1
			runStack(&state, visitChan)
			log.Println("Running:", state.Running)
		}
	}

	channel <- Response{
		Status:  Finished,
		Message: strings.Join(resultPath, " ➡️ "),
	}
}

func SearchBFS_old(start, end string, channel chan Response, forceQuit chan bool) {
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
			time.Sleep(2 * time.Second)
			getLinks([]string{top}, linksChan)
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

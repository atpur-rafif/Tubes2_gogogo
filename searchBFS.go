package main

import (
	"strings"
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
				Next: getLinks(top),
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
	runStack(&state, visitChan)

	channel <- Response{
		Status:  Started,
		Message: "From " + start + " to " + end,
	}

L:
	for {
		select {
		case <-forceQuit:
			return
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
		}
	}

	channel <- Response{
		Status:  Finished,
		Message: strings.Join(resultPath, " ➡️ "),
	}
}

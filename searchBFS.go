package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

const MAX_CONCURRENT = 10

type FetchResult struct {
	From string
	To   []string
}

type StateBFS struct {
	Queue        [][]string
	Visited      map[string]bool
	FetchedCount int // Optimization to start searching for unfetched data
	FetchedData  map[string][]string
	FetchChannel chan FetchResult
	Running      int
}

func (s *StateBFS) prefetch() {
	i := s.FetchedCount
	for i < len(s.Queue) && s.Running < MAX_CONCURRENT {
		path := s.Queue[i]
		current := path[len(path)-1]

		if _, found := s.FetchedData[current]; !found {
			s.Running += 1
			go func() {
				s.FetchChannel <- FetchResult{
					From: current,
					To:   getLinks(current),
				}
			}()
		}
		s.FetchedCount += 1

		i += 1
	}
}

func SearchBFS(start, end string, responseChan chan Response, forceQuit chan bool) {
	shortestpath := 0
	var allSolution [][]string
	responseChan <- Response{
		Status:  Started,
		Message: "Started...",
	}

	s := StateBFS{
		Queue:        make([][]string, 0),
		Visited:      make(map[string]bool),
		FetchedData:  make(map[string][]string),
		FetchChannel: make(chan FetchResult),
		FetchedCount: 0,
		Running:      0,
	}

	s.Queue = append(s.Queue, []string{start})
	s.prefetch()

	var resultPath []string
LO:
	for {
		if len(s.Queue) == 0 {
			log.Println("Path not found")
			break
		}

		path := s.Queue[0]
		s.Queue = s.Queue[1:]
		current := path[len(path)-1]
		s.FetchedCount -= 1
		if s.Visited[current] {
			continue
		}

		for {
			if _, found := s.FetchedData[current]; found {
				s.Visited[current] = true
				break
			}

			select {
			case <-forceQuit:
				return
			case result := <-s.FetchChannel:
				from := result.From
				s.FetchedData[from] = result.To
				s.Running -= 1
				s.prefetch()
			}
		}

		responseChan <- Response{
			Status:  Update,
			Message: "Visited " + current + " with depth " + strconv.Itoa(len(path)-1),
		}

		for _, to := range s.FetchedData[current] {
			if s.Visited[to] {
				continue
			}

			newPath := make([]string, len(path))
			copy(newPath, path)
			newPath = append(newPath, to)
			s.Queue = append(s.Queue, newPath)

			if end == to {
				resultPath = newPath
				shortestpath = len(resultPath)
				allSolution = append(allSolution, resultPath)
			}

			if len(newPath) > shortestpath && len(allSolution) > 0 {
				fmt.Print("Banyak solusi: ")
				fmt.Println(len(allSolution))
				break LO
			}

		}
		s.prefetch()
	}

	responseChan <- Response{
		Status:  Finished,
		Message: strings.Join(resultPath, " ➡️  "),
	}
}

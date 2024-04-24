package main

import (
	"log"
	"strconv"
	"strings"
)

const MAX_CONCURRENT = 10

type FetchResult struct {
	From      string
	Canonical string
	To        []string
}

type StateBFS struct {
	Queue        [][]string
	FetchedCount int // Optimization to start searching for unfetched data
	FetchedData  map[string][]string
	Canonical    map[string]string
	FetchChannel chan FetchResult
	Visited      map[string]bool
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
				canonical, pages := getLinks(current)
				s.FetchChannel <- FetchResult{
					From:      current,
					To:        pages,
					Canonical: canonical,
				}
			}()
		}
		s.FetchedCount += 1

		i += 1
	}
}

func SearchBFS(start, end string, responseChan chan Response, forceQuit chan bool) {
	responseChan <- Response{
		Status:  Started,
		Message: "Started...",
	}

	canonicalEnd, _ := getLinks(end)
	log.Println(canonicalEnd)

	s := StateBFS{
		Queue:        make([][]string, 0),
		FetchedData:  make(map[string][]string),
		FetchChannel: make(chan FetchResult),
		Canonical:    make(map[string]string),
		Visited:      make(map[string]bool),
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
		depth := len(path) - 1
		current := path[depth]
		s.FetchedCount -= 1

		if s.Visited[current] {
			continue
		}

		for {
			if _, found := s.FetchedData[current]; found {
				responseChan <- Response{
					Status:  Update,
					Message: "Visited " + current + " with depth " + strconv.Itoa(len(path)-1),
				}

				canonical := s.Canonical[current]
				s.Visited[current] = true
				s.Visited[canonical] = true

				path[depth] = canonical
				if s.Canonical[current] == canonicalEnd {
					resultPath = path
					break LO
				}

				for _, next := range s.FetchedData[current] {
					newPath := make([]string, len(path))
					copy(newPath, path)
					newPath = append(newPath, next)
					s.Queue = append(s.Queue, newPath)

					if next == canonicalEnd {
						resultPath = newPath
						break LO
					}
				}
				s.prefetch()

				break
			}

			select {
			case <-forceQuit:
				return
			case r := <-s.FetchChannel:
				s.Canonical[r.From] = r.Canonical

				s.FetchedData[r.From] = r.To
				s.FetchedData[r.Canonical] = r.To

				s.Running -= 1
				s.prefetch()
			}
		}
	}

	responseChan <- Response{
		Status:  Finished,
		Message: strings.Join(resultPath, " ➡️  "),
	}
}

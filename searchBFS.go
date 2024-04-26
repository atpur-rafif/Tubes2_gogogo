package main

import (
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
	Start        string
	End          string
	ResultPaths  [][]string
	ResultDepth  int
	Queue        [][]string
	FetchedCount int // Optimization to start searching for unfetched data
	FetchedData  map[string][]string
	Canonical    map[string]string
	FetchChannel chan FetchResult
	Visited      map[string]bool
	Running      int
}

func (s *StateBFS) prefetch() {
	for s.FetchedCount < len(s.Queue) {
		path := s.Queue[s.FetchedCount]
		current := path[len(path)-1]

		if _, found := s.FetchedData[current]; !found {
			if s.Running >= MAX_CONCURRENT {
				if s.FetchedCount == 0 {
					r := <-s.FetchChannel
					s.Canonical[r.From] = r.Canonical
					s.FetchedData[r.Canonical] = r.To

					s.Running -= 1
				} else {
					break
				}
			}

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
	}
}

func SearchBFS(start, end string, responseChan chan Response, forceQuit chan bool) {
	responseChan <- Response{
		Status:  Start,
		Message: "Started...",
	}

	canonicalEnd, _ := getLinks(end)

	s := StateBFS{
		Start:        start,
		End:          canonicalEnd,
		ResultPaths:  make([][]string, 0),
		Queue:        make([][]string, 0),
		FetchedData:  make(map[string][]string),
		FetchChannel: make(chan FetchResult),
		Canonical:    make(map[string]string),
		Visited:      make(map[string]bool),
		FetchedCount: 0,
		Running:      0,
		ResultDepth:  -1,
	}

	s.Queue = append(s.Queue, []string{start})

	// i := 0
	for {
		// if i == 100 {
		// 	break
		// }
		// i += 1

		if len(s.Queue) == 0 {
			break
		}

		s.prefetch()
		path := s.Queue[0]
		s.Queue = s.Queue[1:]
		depth := len(path) - 1
		current := path[depth]
		s.FetchedCount -= 1

		if s.ResultDepth != -1 && depth > s.ResultDepth {
			break
		}

		if canonical, found := s.Canonical[current]; found {
			current = canonical
		}
		if s.Visited[current] {
			continue
		}
		path[depth] = current
		s.Visited[current] = true

		for {
			if _, found := s.FetchedData[current]; found {
				path[depth] = current
				s.Visited[current] = true

				responseChan <- Response{
					Status: Log,
					Message: "Visited article count: " + strconv.Itoa(len(s.FetchedData)) +
						"\nDepth: " + strconv.Itoa(depth) +
						"\nQueue size: " + strconv.Itoa(len(s.Queue)) +
						"\nVisited " + current +
						"\nPath: " + strings.Join(path, " - "),
				}

				var result []string = nil
				if current == canonicalEnd && (s.ResultDepth == -1 || s.ResultDepth == depth) {
					result = path
					s.ResultDepth = depth
				}

				for _, next := range s.FetchedData[current] {
					newPath := make([]string, len(path))
					copy(newPath, path)
					newPath = append(newPath, next)

					if next == canonicalEnd && (s.ResultDepth == -1 || s.ResultDepth == depth+1) {
						result = newPath
						s.ResultDepth = depth + 1
						continue
					}

					s.Queue = append(s.Queue, newPath)
				}

				if result != nil {
					s.ResultPaths = append(s.ResultPaths, result)
					responseChan <- Response{
						Status:  Found,
						Message: result,
					}
				}

				break
			}

			select {
			case <-forceQuit:
				return
			case r := <-s.FetchChannel:
				s.Canonical[r.From] = r.Canonical
				s.FetchedData[r.Canonical] = r.To

				s.Running -= 1
				s.prefetch()

				if current == r.From {
					current = r.Canonical
				}
			}
		}
	}

	responseChan <- Response{
		Status: End,
		// Message: strings.Join(resultPath, " ➡️  "),
	}
}

package main

import (
	"strconv"
	"strings"
)

type StateIDS struct {
	Start string
	End   string

	FetchData     map[string][]string
	FetchedStatus map[string]bool
	Visited       map[string]bool

	FetchChannel chan FetchResult
	AckChannel   chan bool // Acknowledgement channel when traverse thread accept new fetch data

	PrefetcherPath []string
	TraverserPath  []string

	TargetDepth int
	Running     int // Only use this variable on prefetcher thread to avoid race condition

	ResultPath []string
	ForceQuit  bool
}

// TODO: Fix bug Hitler -> Traffic
func prefetcherIDS(s *StateIDS) {
	if s.ForceQuit || s.ResultPath != nil {
		return
	}

	depth := len(s.PrefetcherPath) - 1
	current := s.PrefetcherPath[depth]
	if depth == s.TargetDepth {
		if s.FetchedStatus[current] {
			return
		}

		for s.Running >= MAX_CONCURRENT {
			<-s.AckChannel
			s.Running -= 1
		}

		s.Running += 1
		go func() {
			s.FetchChannel <- FetchResult{
				From: current,
				To:   getLinks(current),
			}
		}()
	} else {
		if !s.FetchedStatus[current] {
			panic("IDS should cached non leaf node, start calling this function from depth zero")
		}

		for _, next := range s.FetchData[current] {
			s.PrefetcherPath = append(s.PrefetcherPath, next)
			prefetcherIDS(s)
			s.PrefetcherPath = s.PrefetcherPath[:len(s.PrefetcherPath)-1]
		}
	}
}

func traverserIDS(s *StateIDS, responseChan chan Response, forceQuit chan bool) {
	if s.ForceQuit || s.ResultPath != nil {
		return
	}

	depth := len(s.TraverserPath) - 1
	current := s.TraverserPath[depth]
	if s.Visited[current] {
		return
	}
	s.Visited[current] = true

	responseChan <- Response{
		Status:  Update,
		Message: "Visited " + current + " with depth " + strconv.Itoa(depth),
	}

	if depth == s.TargetDepth {
		for {
			if s.FetchedStatus[current] {
				break
			}

			select {
			case <-forceQuit:
				s.ForceQuit = true
				return
			case result := <-s.FetchChannel:
				from := result.From
				s.FetchData[from] = result.To
				s.FetchedStatus[from] = true
				s.AckChannel <- true
			}
		}

		for _, next := range s.FetchData[current] {
			if next == s.End {
				s.ResultPath = s.TraverserPath
				s.ResultPath = append(s.TraverserPath, next)
			}
		}

	} else {
		if !s.FetchedStatus[current] {
			panic("IDS should cached non leaf node, start calling this function from depth zero")
		}

		localVisited := make(map[string]bool)
		for _, next := range s.FetchData[current] {
			if s.Visited[next] || localVisited[next] {
				continue
			}

			localVisited[next] = true
			s.TraverserPath = append(s.TraverserPath, next)
			traverserIDS(s, responseChan, forceQuit)
			s.TraverserPath = s.TraverserPath[:len(s.TraverserPath)-1]
		}
	}
}

func SearchIDS(start, end string, responseChan chan Response, forceQuit chan bool) {
	responseChan <- Response{
		Status:  Started,
		Message: "Started...",
	}

	s := StateIDS{
		Start:          start,
		End:            end,
		FetchData:      make(map[string][]string),
		FetchedStatus:  make(map[string]bool),
		Visited:        make(map[string]bool),
		FetchChannel:   make(chan FetchResult),
		AckChannel:     make(chan bool),
		PrefetcherPath: []string{start},
		TraverserPath:  []string{start},
		TargetDepth:    0,
		ResultPath:     nil,
		Running:        0,
	}

	for s.ResultPath == nil {
		s.Visited = make(map[string]bool)
		go func() {
			prefetcherIDS(&s)
			for s.Running > 0 {
				<-s.AckChannel
				s.Running -= 1
			}
		}()
		traverserIDS(&s, responseChan, forceQuit)
		s.TargetDepth += 1
	}

	responseChan <- Response{
		Status:  Finished,
		Message: strings.Join(s.ResultPath, " ➡️  "),
	}
}

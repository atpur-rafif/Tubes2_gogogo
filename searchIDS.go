package main

import (
	"strconv"
	"strings"
	"sync"
)

type StateIDS struct {
	Start      string
	End        string
	ResultPath []string

	Path    []string
	Visited map[string]bool

	FetchedData  map[string][]string
	FetchChannel chan FetchResult
	CurrentFetch []string
	NextFetch    []string

	ForceQuit           bool
	ForceQuitFetch      bool
	ForceQuitFetchMutex sync.Mutex

	MaxDepth int
}

func prefetcherIDS(s *StateIDS) {
	i := 0
	running := 0
	finishedChan := make(chan bool)
	for i < len(s.CurrentFetch) {
		s.ForceQuitFetchMutex.Lock()
		if s.ForceQuitFetch {
			return
		}
		s.ForceQuitFetchMutex.Unlock()

		current := s.CurrentFetch[i]
		if running >= MAX_CONCURRENT {
			<-finishedChan
			running -= 1
		}

		running += 1
		go func() {
			canon, pages := getLinks(current)
			s.FetchChannel <- FetchResult{
				From:      current,
				To:        pages,
				Canonical: canon,
			}
			finishedChan <- true
		}()

		i += 1
	}

	for running > 0 {
		<-finishedChan
		running -= 1
	}
}

func traverserIDS(s *StateIDS, responseChan chan Response, forceQuit chan bool) {
	if s.ForceQuit || s.ResultPath != nil {
		return
	}

	depth := len(s.Path) - 1
	current := s.Path[depth]

	if depth == s.MaxDepth {
		for {
			if pages, found := s.FetchedData[current]; found {
				responseChan <- Response{
					Status:  Update,
					Message: "Visited " + current + " with depth " + strconv.Itoa(depth),
				}

				for _, next := range pages {
					if next == s.End {
						s.ResultPath = s.Path
						s.ResultPath = append(s.ResultPath, next)

						s.ForceQuitFetchMutex.Lock()
						s.ForceQuitFetch = true
						s.ForceQuitFetchMutex.Unlock()
					}

					s.NextFetch = append(s.NextFetch, next)
				}

				break
			}

			select {
			case <-forceQuit:
				s.ForceQuit = true
				return
			case r := <-s.FetchChannel:
				s.FetchedData[r.From] = r.To
			}
		}
	} else {
		if pages, found := s.FetchedData[current]; found {
			for _, next := range pages {
				s.Path = append(s.Path, next)
				traverserIDS(s, responseChan, forceQuit)
				s.Path = s.Path[:len(s.Path)-1]
			}
		} else {
			panic("Non-leaf node should be cached in previous iteration, call this function from depth 0")
		}
	}
}

func SearchIDS(start, end string, responseChan chan Response, forceQuit chan bool) {
	responseChan <- Response{
		Status:  Started,
		Message: "Started...",
	}

	s := StateIDS{
		Start:        start,
		End:          end,
		ResultPath:   nil,
		Path:         make([]string, 0),
		Visited:      make(map[string]bool),
		FetchedData:  make(map[string][]string),
		FetchChannel: make(chan FetchResult),
		CurrentFetch: make([]string, 0),
		NextFetch:    make([]string, 0),
		MaxDepth:     0,
	}

	s.Path = append(s.Path, s.Start)
	s.NextFetch = append(s.NextFetch, s.Start)

	for s.ResultPath == nil {
		s.CurrentFetch = make([]string, 0)
		for _, nextFetch := range s.NextFetch {
			if _, found := s.FetchedData[nextFetch]; !found {
				s.CurrentFetch = append(s.CurrentFetch, nextFetch)
			}
		}

		s.NextFetch = make([]string, 0)
		go prefetcherIDS(&s)
		traverserIDS(&s, responseChan, forceQuit)
		if s.ForceQuit {
			return
		}

		s.MaxDepth += 1
	}

	responseChan <- Response{
		Status:  Finished,
		Message: strings.Join(s.ResultPath, " ➡️  "),
	}
}

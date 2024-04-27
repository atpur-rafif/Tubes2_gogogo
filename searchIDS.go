package main

import (
	"strconv"
	"strings"
	"sync"
)

type StateIDS struct {
	Start        string
	End          string
	CanonicalEnd string
	ResultPaths  [][]string
	Path         []string
	PathSet      map[string]bool
	Canonical    map[string]string

	FetchedData  map[string][]string
	FetchChannel chan FetchResult
	CurrentFetch []string
	NextFetch    []string

	ForceQuit           bool
	ForceQuitFetch      bool
	ForceQuitFetchMutex sync.Mutex

	MaxDepth    int
	ResultDepth int
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
	depth := len(s.Path) - 1
	current := s.Path[depth]

	if canonical, found := s.Canonical[current]; found {
		current = canonical
	}

	if depth == s.MaxDepth {
		for {
			if pages, found := s.FetchedData[current]; found {
				s.Path[depth] = current

				statusLog := ""
				if s.ResultDepth == depth {
					statusLog = "\nValidating " + current
				} else {
					statusLog = "\nTraversed: " + strings.Join(s.Path, " - ")
				}

				responseChan <- Response{
					Status: Log,
					Message: "Visited article count: " + strconv.Itoa(len(s.FetchedData)) +
						"\nIteration: " + strconv.Itoa(s.MaxDepth) +
						"\nDepth: " + strconv.Itoa(depth) +
						statusLog,
				}

				var result []string = nil
				if current == s.CanonicalEnd && (s.ResultDepth == -1 || s.ResultDepth == depth) {
					result = s.Path
					s.ResultDepth = depth
				}

				if s.ResultDepth == -1 || depth <= s.ResultDepth {
					for _, next := range pages {
						if next == s.CanonicalEnd && (s.ResultDepth == -1 || s.ResultDepth == depth+1) {
							path := make([]string, len(s.Path))
							copy(path, s.Path)
							path = append(path, next)

							result = path
							s.ResultDepth = depth + 1
							continue
						}

						s.NextFetch = append(s.NextFetch, next)
					}
				}

				if result != nil {
					responseChan <- Response{
						Status:  Found,
						Message: result,
					}

					s.ResultPaths = append(s.ResultPaths, result)
					return
				}

				break
			}

			select {
			case <-forceQuit:
				s.ForceQuit = true
				return
			case r := <-s.FetchChannel:
				s.FetchedData[r.Canonical] = r.To
				s.Canonical[r.From] = r.Canonical

				if r.From == current {
					current = r.Canonical
				}
			}
		}
	} else {
		if pages, found := s.FetchedData[current]; found {
			for _, next := range pages {
				if canonical, found := s.Canonical[next]; found {
					next = canonical
				}

				if s.PathSet[next] {
					continue
				}

				s.Path = append(s.Path, next)
				s.PathSet[next] = true
				traverserIDS(s, responseChan, forceQuit)
				delete(s.PathSet, next)
				s.Path = s.Path[:len(s.Path)-1]
			}
		} else {
			panic("Non-leaf node should be cached in previous iteration, call this function from depth 0")
		}
	}
}

func SearchIDS(start, end string, responseChan chan Response, forceQuit chan bool) {
	responseChan <- Response{
		Status:  Start,
		Message: "Started...",
	}

	s := StateIDS{
		Start:        start,
		End:          end,
		ResultPaths:  make([][]string, 0),
		Path:         make([]string, 0),
		PathSet:      make(map[string]bool),
		Canonical:    make(map[string]string),
		FetchedData:  make(map[string][]string),
		FetchChannel: make(chan FetchResult),
		CurrentFetch: make([]string, 0),
		NextFetch:    make([]string, 0),
		MaxDepth:     0,
		ResultDepth:  -1,
	}

	s.Path = append(s.Path, s.Start)
	s.NextFetch = append(s.NextFetch, s.Start)
	canonicalEnd, _ := getLinks(end)
	s.CanonicalEnd = canonicalEnd

	for s.ResultDepth == -1 || s.MaxDepth <= s.ResultDepth {
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
		Status:  End,
		Message: "Search finished",
	}
}

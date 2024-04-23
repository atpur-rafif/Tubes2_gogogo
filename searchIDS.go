package main

import "log"

type StateIDS struct {
	Start string
	End   string

	Path    []string
	Visited map[string]bool

	FetchedData  map[string][]string
	FetchChannel chan FetchResult
	CurrentFetch []string
	NextFetch    []string

	MaxDepth int
}

func prefetcherIDS(s *StateIDS) {
	i := 0
	running := 0
	finishedChan := make(chan bool)
	for i < len(s.CurrentFetch) {
		if running >= MAX_CONCURRENT {
			<-finishedChan
			running -= 1
		}

		current := s.CurrentFetch[i]
		running += 1
		go func() {
			s.FetchChannel <- FetchResult{
				From: current,
				To:   getLinks(current),
			}
			finishedChan <- true
		}()

		i += 1
	}
}

func traverserIDS(s *StateIDS, responseChan chan Response, forceQuit chan bool) {
	depth := len(s.Path) - 1
	current := s.Path[len(s.Path)-1]
	log.Println(s.Path)

	// if s.Visited[current] {
	// 	return
	// }
	s.Visited[current] = true

	if depth == s.MaxDepth {
		for {
			if _, found := s.FetchedData[current]; found {
				break
			}

			select {
			case <-forceQuit:
				return
			case result := <-s.FetchChannel:
				from := result.From
				s.FetchedData[from] = result.To
			}
		}

		for _, next := range s.FetchedData[current] {
			// if s.Visited[next] {
			// 	continue
			// }

			if _, found := s.FetchedData[next]; !found {
				s.NextFetch = append(s.NextFetch, next)
			}

			if next == s.End {
				log.Println(s.Path, next)
				panic("FOUND")
			}
		}
	} else {
		for _, next := range s.FetchedData[current] {
			// if s.Visited[next] {
			// 	continue
			// }

			s.Path = append(s.Path, next)
			traverserIDS(s, responseChan, forceQuit)
			s.Path = s.Path[:len(s.Path)-1]
		}
	}
}

func SearchIDS(start, end string, responseChan chan Response, forceQuit chan bool) {
	s := StateIDS{
		Start:        start,
		End:          end,
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

	for i := 0; i < 10; i += 1 {
		s.CurrentFetch = s.NextFetch
		s.NextFetch = make([]string, 0)
		s.Visited = make(map[string]bool)
		go prefetcherIDS(&s)
		traverserIDS(&s, responseChan, forceQuit)
		log.Println("Finished depth", s.MaxDepth)
		s.MaxDepth += 1
	}
}

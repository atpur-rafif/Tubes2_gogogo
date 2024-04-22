package main

import (
	"strings"
)

type StateIDS struct {
	Pages       map[string][]string
	Start       string
	End         string
	ForceQuit   bool
	ResultPaths [][]string
}

// Only one data accross thread, and access this data only on main thread to avoid data race
type GlobalStateDLS struct {
	Running   int
	MaxDepth  int
	PathFound bool
	Visited   map[string]bool
	StateIDS  *StateIDS
}

// Create one every branching happen
type StateDLS struct {
	Path   []string
	Global *GlobalStateDLS
}

// Save current StateDLS to defer processing until fetch finished
type BranchDLS struct {
	Path []string
	Next []string
}

func DLS(stateIDS *StateIDS, maxDepth int, responseChan chan Response, forceQuit chan bool) bool {
	global := GlobalStateDLS{
		Running:   0,
		PathFound: false,
		MaxDepth:  maxDepth,
		StateIDS:  stateIDS,
		Visited:   make(map[string]bool),
	}

	startState := StateDLS{
		Path:   []string{stateIDS.Start},
		Global: &global,
	}

	branchChan := make(chan BranchDLS)
	recurDLS(&startState, responseChan, branchChan, forceQuit)
	for global.Running != 0 {
		waitBranch(&global, responseChan, branchChan, forceQuit)
	}

	return global.PathFound
}

func branching(s *StateDLS, _ chan Response, branchChan chan BranchDLS, _ chan bool) {
	copyPath := make([]string, len(s.Path))
	copy(copyPath, s.Path)
	s.Global.Running += 1
	go func() {
		current := copyPath[len(copyPath)-1]
		branchChan <- BranchDLS{
			Path: copyPath,
			Next: getLinks(current),
		}
	}()
}

func waitBranch(s *GlobalStateDLS, responseChan chan Response, branchChan chan BranchDLS, forceQuit chan bool) {
	select {
	case b := <-branchChan:
		s.Running -= 1
		current := b.Path[len(b.Path)-1]
		stateDLS := StateDLS{
			Path:   b.Path,
			Global: s,
		}
		s.StateIDS.Pages[current] = b.Next
		recurDLS(&stateDLS, responseChan, branchChan, forceQuit)
	case <-forceQuit:
		s.StateIDS.ForceQuit = true
	}
}

func recurDLS(s *StateDLS, responseChan chan Response, branchChan chan BranchDLS, forceQuit chan bool) {
	current := s.Path[len(s.Path)-1]

	if len(s.Path) == s.Global.MaxDepth+1 {
		if current == s.Global.StateIDS.End {
			s.Global.PathFound = true
			s.Global.StateIDS.ResultPaths = append(s.Global.StateIDS.ResultPaths, s.Path)
			responseChan <- Response{
				Status:  Finished,
				Message: strings.Join(s.Path, " ➡️  "),
			}
			go func() {
				forceQuit <- true
			}()
		}
		return
	}

	// Page already fetched
	if pages, found := s.Global.StateIDS.Pages[current]; found {
		if !s.Global.StateIDS.ForceQuit {
			responseChan <- Response{
				Status:  Update,
				Message: "Visited " + current,
			}
		}

		s.Global.Visited[current] = true
		nextIterated := make(map[string]bool)
		for _, next := range pages {
			if s.Global.Visited[next] || nextIterated[next] {
				continue
			}

			nextIterated[next] = true
			s.Path = append(s.Path, next)
			recurDLS(s, responseChan, branchChan, forceQuit)
			s.Path = s.Path[:len(s.Path)-1]
		}
		return
	}

	for s.Global.Running >= MAX_CONCURRENT {
		waitBranch(s.Global, responseChan, branchChan, forceQuit)
	}

	if s.Global.StateIDS.ForceQuit {
		return
	}
	branching(s, responseChan, branchChan, forceQuit)
}

func SearchIDS(start, end string, responseChan chan Response, forceQuit chan bool) {
	stateIDS := StateIDS{
		ForceQuit: false,
		Pages:     make(map[string][]string, 0),
		Start:     start,
		End:       end,
	}

	depth := 0
	for !DLS(&stateIDS, depth, responseChan, forceQuit) {
		depth += 1
	}
}

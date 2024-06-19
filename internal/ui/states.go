package ui

import "sync"

type State struct {
	mu      sync.RWMutex
	current uiState
}

type uiState int

func (s *State) Set(v uiState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.current = v

}
func (s *State) Get() uiState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}

const (
	SelectFile uiState = iota
	Running
	Stale
)

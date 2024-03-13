package cli

import "sync"

type State struct {
	mu      sync.RWMutex
	current uiState
}

type uiState int

func (this *State) Set(v uiState) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.current = v

}
func (this *State) Get() uiState {
	this.mu.RLock()
	defer this.mu.RUnlock()
	return this.current
}

const (
	SelectFile uiState = iota
	Running
	Stale
)

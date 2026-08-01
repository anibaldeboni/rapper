package config

import (
	"sync"
)

type runManager struct {
	requestMgr *requestManager
	listeners  []func(*Config)
	mu         sync.RWMutex
}

func NewRunManager(dir string) (*runManager, error) {
	loader := NewLoader()
	reqMgr := newRequestManager(loader)

	_, err := reqMgr.discover(dir)
	if err != nil {
		return nil, err
	}

	return &runManager{
		requestMgr: reqMgr,
		listeners:  make([]func(*Config), 0),
	}, nil
}

func (m *runManager) Get() *Config {
	return m.requestMgr.getActiveConfig()
}

func (m *runManager) GetRequest(name string) *Config {
	profile := m.requestMgr.getByName(name)
	if profile == nil {
		return nil
	}
	return profile.Config
}

func (m *runManager) Update(cfg *Config) error {
	if err := m.updateInMemory(cfg); err != nil {
		return err
	}
	m.notify(cfg)
	return nil
}

func (m *runManager) updateInMemory(cfg *Config) error {
	return m.requestMgr.updateActive(cfg)
}

func (m *runManager) notify(cfg *Config) {
	m.mu.RLock()
	listeners := make([]func(*Config), len(m.listeners))
	copy(listeners, m.listeners)
	m.mu.RUnlock()

	for _, fn := range listeners {
		fn(cfg)
	}
}

func (m *runManager) Save() error {
	return m.requestMgr.save()
}

func (m *runManager) OnChange(callback func(*Config)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.listeners = append(m.listeners, callback)
}

func (m *runManager) ListRequests() []string {
	profiles := m.requestMgr.list()
	names := make([]string, len(profiles))
	for i, p := range profiles {
		names[i] = p.Name
	}
	return names
}

func (m *runManager) GetActiveRequest() string {
	active := m.requestMgr.getActive()
	if active == nil {
		return ""
	}
	return active.Name
}

func (m *runManager) SetActiveRequest(name string) error {
	if err := m.requestMgr.setActive(name); err != nil {
		return err
	}

	cfg := m.Get()

	m.mu.RLock()
	listeners := make([]func(*Config), len(m.listeners))
	copy(listeners, m.listeners)
	m.mu.RUnlock()

	for _, fn := range listeners {
		fn(cfg)
	}

	return nil
}

package config

import (
	"sync"
)

type managerImpl struct {
	profileMgr *profileManagerImpl
	listeners  []func(*Config)
	mu         sync.RWMutex
}

func NewManager(dir string) (*managerImpl, error) {
	loader := NewLoader()
	profileMgr := newProfileManager(loader)

	_, err := profileMgr.discover(dir)
	if err != nil {
		return nil, err
	}

	return &managerImpl{
		profileMgr: profileMgr,
		listeners:  make([]func(*Config), 0),
	}, nil
}

func (m *managerImpl) Get() *Config {
	return m.profileMgr.getActiveConfig()
}

func (m *managerImpl) GetProfile(name string) *Config {
	profile := m.profileMgr.getByName(name)
	if profile == nil {
		return nil
	}
	return profile.Config
}

func (m *managerImpl) Update(cfg *Config) error {
	if err := m.updateInMemory(cfg); err != nil {
		return err
	}
	m.notify(cfg)
	return nil
}

func (m *managerImpl) updateInMemory(cfg *Config) error {
	return m.profileMgr.updateActive(cfg)
}

func (m *managerImpl) notify(cfg *Config) {
	m.mu.RLock()
	listeners := make([]func(*Config), len(m.listeners))
	copy(listeners, m.listeners)
	m.mu.RUnlock()

	for _, fn := range listeners {
		fn(cfg)
	}
}

func (m *managerImpl) Save() error {
	return m.profileMgr.save()
}

func (m *managerImpl) OnChange(callback func(*Config)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.listeners = append(m.listeners, callback)
}

func (m *managerImpl) ListProfiles() []string {
	profiles := m.profileMgr.list()
	names := make([]string, len(profiles))
	for i, p := range profiles {
		names[i] = p.Name
	}
	return names
}

func (m *managerImpl) GetActiveProfile() string {
	active := m.profileMgr.getActive()
	if active == nil {
		return ""
	}
	return active.Name
}

func (m *managerImpl) SetActiveProfile(name string) error {
	if err := m.profileMgr.setActive(name); err != nil {
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

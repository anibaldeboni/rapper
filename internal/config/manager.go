package config

import (
	"sync"
)

// Manager manages configuration with hot-reload support
//
//go:generate mockgen -destination mock/manager_mock.go github.com/anibaldeboni/rapper/internal/config Manager
type Manager interface {
	// Get returns the current configuration
	Get() *Config

	// Update updates the current configuration in memory
	Update(cfg *Config) error

	// Save persists the current configuration to disk
	Save() error

	// OnChange registers a callback to be notified when configuration changes
	OnChange(callback func(*Config))

	// GetProfileManager returns the underlying ProfileManager
	GetProfileManager() ProfileManager
}

type managerImpl struct {
	profileMgr ProfileManager
	listeners  []func(*Config)
	mu         sync.RWMutex
}

// NewManager creates a new Manager instance
// It discovers all .yml files in the specified directory and loads them as profiles
func NewManager(dir string) (Manager, error) {
	loader := NewLoader()
	profileMgr := NewProfileManager(loader)

	// Discover profiles in the directory
	_, err := profileMgr.Discover(dir)
	if err != nil {
		return nil, err
	}

	return &managerImpl{
		profileMgr: profileMgr,
		listeners:  make([]func(*Config), 0),
	}, nil
}

// Get returns the current configuration from the active profile
func (m *managerImpl) Get() *Config {
	active := m.profileMgr.GetActive()
	if active == nil {
		return nil
	}
	return active.Config
}

// Update updates the current configuration and notifies all listeners
func (m *managerImpl) Update(cfg *Config) error {
	if err := m.profileMgr.UpdateActive(cfg); err != nil {
		return err
	}

	// Notify all registered listeners
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, listener := range m.listeners {
		listener(cfg)
	}

	return nil
}

// Save persists the current configuration to the active profile's YAML file
func (m *managerImpl) Save() error {
	return m.profileMgr.Save()
}

// OnChange registers a callback to be notified when configuration changes
func (m *managerImpl) OnChange(callback func(*Config)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.listeners = append(m.listeners, callback)
}

// GetProfileManager returns the underlying ProfileManager
func (m *managerImpl) GetProfileManager() ProfileManager {
	return m.profileMgr
}

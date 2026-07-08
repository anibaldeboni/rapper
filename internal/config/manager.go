package config

import (
	"sync"
)

type managerImpl struct {
	profileMgr *profileManagerImpl
	listeners  []func(*Config)
	mu         sync.RWMutex
}

// NewManager creates a new Manager instance.
// It discovers all .yml files in the specified directory and loads them as profiles.
func NewManager(dir string) (*managerImpl, error) {
	loader := NewLoader()
	profileMgr := newProfileManager(loader)

	// Discover profiles in the directory
	_, err := profileMgr.discover(dir)
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
	active := m.profileMgr.getActive()
	if active == nil {
		return nil
	}
	return active.Config
}

// GetProfile returns the configuration of the named profile, or nil
// if no profile with that name exists. The settings view uses this
// to preview a non-active profile's config in the form fields before
// the user decides to activate it with SetActiveProfile.
func (m *managerImpl) GetProfile(name string) *Config {
	profile := m.profileMgr.getByName(name)
	if profile == nil {
		return nil
	}
	return profile.Config
}

// Update updates the current configuration and notifies all listeners
func (m *managerImpl) Update(cfg *Config) error {
	if err := m.profileMgr.updateActive(cfg); err != nil {
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
	return m.profileMgr.save()
}

// OnChange registers a callback to be notified when configuration changes
func (m *managerImpl) OnChange(callback func(*Config)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.listeners = append(m.listeners, callback)
}

// ListProfiles returns names of all available profiles
func (m *managerImpl) ListProfiles() []string {
	profiles := m.profileMgr.list()
	names := make([]string, len(profiles))
	for i, p := range profiles {
		names[i] = p.Name
	}
	return names
}

// GetActiveProfile returns the name of the active profile
func (m *managerImpl) GetActiveProfile() string {
	active := m.profileMgr.getActive()
	if active == nil {
		return ""
	}
	return active.Name
}

// SetActiveProfile switches to the specified profile and notifies all
// OnChange listeners with the new active *Config, mirroring the notification
// block in Update.
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

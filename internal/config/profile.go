package config

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
)

// Profile represents a configuration profile
type Profile struct {
	Name     string  // Name of the profile (e.g., "api1", "production")
	FilePath string  // Path to the YAML file (e.g., "./api1.yml")
	Config   *Config // Loaded configuration
}

// ProfileManager manages multiple configuration profiles
//
//go:generate mockgen -destination mock/profile_mock.go github.com/anibaldeboni/rapper/internal/config ProfileManager
type ProfileManager interface {
	// Discover finds all .yml and .yaml files in the directory
	Discover(dir string) ([]Profile, error)

	// List returns all available profiles
	List() []Profile

	// GetActive returns the currently active profile
	GetActive() *Profile

	// SetActive switches to a different profile by name
	SetActive(name string) error

	// Save persists the active profile to its YAML file
	Save() error

	// UpdateActive updates the configuration of the active profile
	UpdateActive(cfg *Config) error
}

type profileManagerImpl struct {
	profiles     []Profile
	activeIndex  int
	configLoader *Loader
	mu           sync.RWMutex
}

// NewProfileManager creates a new ProfileManager instance
func NewProfileManager(loader *Loader) ProfileManager {
	return &profileManagerImpl{
		profiles:     make([]Profile, 0),
		activeIndex:  0,
		configLoader: loader,
	}
}

// Discover finds all .yml and .yaml files in the directory and loads them as profiles
func (pm *profileManagerImpl) Discover(dir string) ([]Profile, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Search for .yml files
	ymlFiles, err := filepath.Glob(filepath.Join(dir, "*.yml"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob yml files: %w", err)
	}

	// Search for .yaml files
	yamlFiles, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob yaml files: %w", err)
	}

	// Combine both lists
	files := append(ymlFiles, yamlFiles...)

	if len(files) == 0 {
		return nil, fmt.Errorf("no .yml or .yaml files found in %s", dir)
	}

	profiles := make([]Profile, 0, len(files))

	for _, filePath := range files {
		// Load the configuration file
		cfg, err := pm.configLoader.Load(filePath)
		if err != nil {
			// Skip invalid files but log the error
			log.Printf("Skipping invalid config file %s: %v", filePath, err)
			continue
		}

		// Extract profile name from filename (without extension)
		baseName := filepath.Base(filePath)
		name := strings.TrimSuffix(baseName, filepath.Ext(baseName))

		profiles = append(profiles, Profile{
			Name:     name,
			FilePath: filePath,
			Config:   cfg,
		})
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("no valid config files found")
	}

	pm.profiles = profiles
	pm.activeIndex = 0 // First profile is active by default

	return profiles, nil
}

// List returns all available profiles
func (pm *profileManagerImpl) List() []Profile {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Return a copy to prevent external modification
	profiles := make([]Profile, len(pm.profiles))
	copy(profiles, pm.profiles)
	return profiles
}

// GetActive returns the currently active profile
func (pm *profileManagerImpl) GetActive() *Profile {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.activeIndex < 0 || pm.activeIndex >= len(pm.profiles) {
		return nil
	}

	// Return a pointer to the profile (not a copy)
	return &pm.profiles[pm.activeIndex]
}

// SetActive switches to a different profile by name
func (pm *profileManagerImpl) SetActive(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for i, profile := range pm.profiles {
		if profile.Name == name {
			pm.activeIndex = i
			return nil
		}
	}

	return fmt.Errorf("profile %s not found", name)
}

// UpdateActive updates the configuration of the active profile
func (pm *profileManagerImpl) UpdateActive(cfg *Config) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.activeIndex < 0 || pm.activeIndex >= len(pm.profiles) {
		return fmt.Errorf("no active profile")
	}

	pm.profiles[pm.activeIndex].Config = cfg
	return nil
}

// Save persists the active profile to its YAML file
func (pm *profileManagerImpl) Save() error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.activeIndex < 0 || pm.activeIndex >= len(pm.profiles) {
		return fmt.Errorf("no active profile")
	}

	active := &pm.profiles[pm.activeIndex]

	// Save to the original YAML file
	err := pm.configLoader.Save(active.FilePath, active.Config)
	if err != nil {
		return fmt.Errorf("failed to save profile %s: %w", active.Name, err)
	}

	return nil
}

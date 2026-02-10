package config

import (
	"errors"
	"fmt"
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

// profileManagerImpl manages multiple configuration profiles.
// This is an internal implementation - not exposed as an interface.
type profileManagerImpl struct {
	profiles     []Profile
	activeIndex  int
	configLoader *Loader
	mu           sync.RWMutex
}

// newProfileManager creates a new profile manager instance
func newProfileManager(loader *Loader) *profileManagerImpl {
	return &profileManagerImpl{
		profiles:     make([]Profile, 0),
		activeIndex:  0,
		configLoader: loader,
	}
}

// discover finds all .yml and .yaml files in the directory and loads them as profiles
func (pm *profileManagerImpl) discover(dir string) ([]Profile, error) {
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
		// Extract filename to check if it should be skipped
		baseName := filepath.Base(filePath)

		// Skip hidden files (starting with .) and known non-profile config files
		if strings.HasPrefix(baseName, ".") {
			continue
		}

		// Load the configuration file
		cfg, err := pm.configLoader.Load(filePath)
		if err != nil {
			// Silently skip invalid files (they might be config files for other tools)
			continue
		}

		// Extract profile name from filename (without extension)
		name := strings.TrimSuffix(baseName, filepath.Ext(baseName))

		profiles = append(profiles, Profile{
			Name:     name,
			FilePath: filePath,
			Config:   cfg,
		})
	}

	if len(profiles) == 0 {
		return nil, errors.New("no valid config files found")
	}

	pm.profiles = profiles
	pm.activeIndex = 0 // First profile is active by default

	return profiles, nil
}

// list returns all available profiles
func (pm *profileManagerImpl) list() []Profile {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Return a copy to prevent external modification
	profiles := make([]Profile, len(pm.profiles))
	copy(profiles, pm.profiles)
	return profiles
}

// getActive returns the currently active profile
func (pm *profileManagerImpl) getActive() *Profile {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.activeIndex < 0 || pm.activeIndex >= len(pm.profiles) {
		return nil
	}

	// Return a pointer to the profile (not a copy)
	return &pm.profiles[pm.activeIndex]
}

// setActive switches to a different profile by name
func (pm *profileManagerImpl) setActive(name string) error {
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

// updateActive updates the configuration of the active profile
func (pm *profileManagerImpl) updateActive(cfg *Config) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.activeIndex < 0 || pm.activeIndex >= len(pm.profiles) {
		return errors.New("no active profile")
	}

	pm.profiles[pm.activeIndex].Config = cfg
	return nil
}

// save persists the active profile to its YAML file
func (pm *profileManagerImpl) save() error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.activeIndex < 0 || pm.activeIndex >= len(pm.profiles) {
		return errors.New("no active profile")
	}

	active := &pm.profiles[pm.activeIndex]

	// Save to the original YAML file
	err := pm.configLoader.Save(active.FilePath, active.Config)
	if err != nil {
		return fmt.Errorf("failed to save profile %s: %w", active.Name, err)
	}

	return nil
}

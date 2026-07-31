package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

type Profile struct {
	Name     string
	FilePath string
	Config   *Config
}

type profileManagerImpl struct {
	profiles     []Profile
	activeIndex  int
	configLoader *Loader
	mu           sync.RWMutex
}

func newProfileManager(loader *Loader) *profileManagerImpl {
	return &profileManagerImpl{
		profiles:     make([]Profile, 0),
		activeIndex:  0,
		configLoader: loader,
	}
}

func (pm *profileManagerImpl) discover(dir string) ([]Profile, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	ymlFiles, err := filepath.Glob(filepath.Join(dir, "*.yml"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob yml files: %w", err)
	}

	yamlFiles, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob yaml files: %w", err)
	}

	files := append(ymlFiles, yamlFiles...)

	if len(files) == 0 {
		return nil, fmt.Errorf("no .yml or .yaml files found in %s", dir)
	}

	profiles := make([]Profile, 0, len(files))

	for _, filePath := range files {

		baseName := filepath.Base(filePath)

		if strings.HasPrefix(baseName, ".") {
			continue
		}

		cfg, err := pm.configLoader.Load(filePath)
		if err != nil {

			continue
		}

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
	pm.activeIndex = 0

	return profiles, nil
}

func (pm *profileManagerImpl) list() []Profile {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	profiles := make([]Profile, len(pm.profiles))
	copy(profiles, pm.profiles)
	return profiles
}

func (pm *profileManagerImpl) getActive() *Profile {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.activeIndex < 0 || pm.activeIndex >= len(pm.profiles) {
		return nil
	}

	return &pm.profiles[pm.activeIndex]
}

func (pm *profileManagerImpl) getActiveConfig() *Config {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.activeIndex < 0 || pm.activeIndex >= len(pm.profiles) {
		return nil
	}
	return pm.profiles[pm.activeIndex].Config
}

func (pm *profileManagerImpl) getByName(name string) *Profile {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for i := range pm.profiles {
		if pm.profiles[i].Name == name {
			return &pm.profiles[i]
		}
	}
	return nil
}

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

func (pm *profileManagerImpl) updateActive(cfg *Config) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.activeIndex < 0 || pm.activeIndex >= len(pm.profiles) {
		return errors.New("no active profile")
	}

	pm.profiles[pm.activeIndex].Config = cfg
	return nil
}

func (pm *profileManagerImpl) updateByName(name string, cfg *Config) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for i := range pm.profiles {
		if pm.profiles[i].Name == name {
			pm.profiles[i].Config = cfg
			return nil
		}
	}
	return nil
}

func (pm *profileManagerImpl) save() error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.activeIndex < 0 || pm.activeIndex >= len(pm.profiles) {
		return errors.New("no active profile")
	}

	active := &pm.profiles[pm.activeIndex]

	err := pm.configLoader.Save(active.FilePath, active.Config)
	if err != nil {
		return fmt.Errorf("failed to save profile %s: %w", active.Name, err)
	}

	return nil
}

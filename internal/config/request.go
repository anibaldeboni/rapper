package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

type Request struct {
	Name     string
	FilePath string
	Config   *Config
}

type requestManager struct {
	requests     []Request
	activeIndex  int
	configLoader *Loader
	mu           sync.RWMutex
}

func newRequestManager(loader *Loader) *requestManager {
	return &requestManager{
		requests:     make([]Request, 0),
		activeIndex:  0,
		configLoader: loader,
	}
}

func (rm *requestManager) discover(dir string) ([]Request, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

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

	profiles := make([]Request, 0, len(files))

	for _, filePath := range files {

		baseName := filepath.Base(filePath)

		if strings.HasPrefix(baseName, ".") {
			continue
		}

		cfg, err := rm.configLoader.Load(filePath)
		if err != nil {

			continue
		}

		name := strings.TrimSuffix(baseName, filepath.Ext(baseName))

		profiles = append(profiles, Request{
			Name:     name,
			FilePath: filePath,
			Config:   cfg,
		})
	}

	if len(profiles) == 0 {
		return nil, errors.New("no valid config files found")
	}

	rm.requests = profiles
	rm.activeIndex = 0

	return profiles, nil
}

func (rm *requestManager) list() []Request {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	profiles := make([]Request, len(rm.requests))
	copy(profiles, rm.requests)
	return profiles
}

func (rm *requestManager) getActive() *Request {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if rm.activeIndex < 0 || rm.activeIndex >= len(rm.requests) {
		return nil
	}

	return &rm.requests[rm.activeIndex]
}

func (rm *requestManager) getActiveConfig() *Config {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if rm.activeIndex < 0 || rm.activeIndex >= len(rm.requests) {
		return nil
	}
	return rm.requests[rm.activeIndex].Config
}

func (rm *requestManager) getByName(name string) *Request {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	for i := range rm.requests {
		if rm.requests[i].Name == name {
			return &rm.requests[i]
		}
	}
	return nil
}

func (rm *requestManager) setActive(name string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for i, profile := range rm.requests {
		if profile.Name == name {
			rm.activeIndex = i
			return nil
		}
	}

	return fmt.Errorf("profile %s not found", name)
}

func (rm *requestManager) updateActive(cfg *Config) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.activeIndex < 0 || rm.activeIndex >= len(rm.requests) {
		return errors.New("no active profile")
	}

	rm.requests[rm.activeIndex].Config = cfg
	return nil
}

func (rm *requestManager) updateByName(name string, cfg *Config) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for i := range rm.requests {
		if rm.requests[i].Name == name {
			rm.requests[i].Config = cfg
			return nil
		}
	}
	return nil
}

func (rm *requestManager) save() error {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if rm.activeIndex < 0 || rm.activeIndex >= len(rm.requests) {
		return errors.New("no active profile")
	}

	active := &rm.requests[rm.activeIndex]

	err := rm.configLoader.Save(active.FilePath, active.Config)
	if err != nil {
		return fmt.Errorf("failed to save profile %s: %w", active.Name, err)
	}

	return nil
}

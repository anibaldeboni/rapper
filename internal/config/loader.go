package config

import (
	"errors"
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v3"
)

// Loader handles loading and parsing YAML configuration files
type Loader struct{}

// NewLoader creates a new Loader instance
func NewLoader() *Loader {
	return &Loader{}
}

// Load reads and parses a new-format YAML configuration file.
func (l *Loader) Load(filePath string) (*Config, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	var config Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", filePath, err)
	}

	if err := l.validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config in %s: %w", filePath, err)
	}

	return &config, nil
}

// validateConfig validates the configuration structure
func (l *Loader) validateConfig(cfg *Config) error {
	if cfg.Request.Method == "" {
		return errors.New("request.method is required")
	}
	if cfg.Request.URLTemplate == "" {
		return errors.New("request.url_template is required")
	}
	if len(cfg.CSV.Fields) == 0 {
		return errors.New("csv.fields is required")
	}
	if cfg.Workers < 0 {
		return errors.New("workers must be >= 0")
	}
	return nil
}

// Save writes a configuration to a YAML file
func (l *Loader) Save(filePath string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file %s: %w", filePath, err)
	}

	return nil
}

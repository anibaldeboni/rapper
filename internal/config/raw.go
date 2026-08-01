package config

import (
	"errors"
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v3"
)

var ErrEmptyContent = errors.New("config content is empty")

func parseConfigFile(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if vErr := validateConfigShim(&cfg); vErr != nil {
		return nil, fmt.Errorf("invalid config: %w", vErr)
	}
	return &cfg, nil
}

func validateConfigShim(cfg *Config) error {
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

func (m *runManager) GetRequestRaw(name string) (string, error) {
	profile := m.requestMgr.getByName(name)
	if profile == nil {
		return "", fmt.Errorf("profile %q not found", name)
	}
	data, err := os.ReadFile(profile.FilePath)
	if err != nil {
		return "", fmt.Errorf("read profile %q (%s): %w", name, profile.FilePath, err)
	}
	return string(data), nil
}

func (m *runManager) SaveRequestRaw(name string, content string) error {
	if content == "" {
		return fmt.Errorf("save profile %q: %w", name, ErrEmptyContent)
	}
	profile := m.requestMgr.getByName(name)
	if profile == nil {
		return fmt.Errorf("profile %q not found", name)
	}

	cfg, err := parseConfigFile([]byte(content))
	if err != nil {
		return fmt.Errorf("save profile %q: invalid YAML: %w", name, err)
	}

	if err := os.WriteFile(profile.FilePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write profile %q (%s): %w", name, profile.FilePath, err)
	}

	if err := m.requestMgr.updateByName(name, cfg); err != nil {
		return fmt.Errorf("update in-memory config for %q: %w", name, err)
	}

	if name == m.GetActiveRequest() {
		m.notify(cfg)
	}

	return nil
}

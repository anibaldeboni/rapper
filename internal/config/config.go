package config

import (
	"fmt"
	"os"

	"github.com/anibaldeboni/rapper/internal/styles"
	"github.com/anibaldeboni/rapper/internal/utils"
	yaml "gopkg.in/yaml.v3"
)

// CSVConfig holds CSV-specific configuration
type CSVConfig struct {
	Separator string   `yaml:"separator"`
	Fields    []string `yaml:"fields"`
}

// RequestConfig holds HTTP request configuration
type RequestConfig struct {
	Method       string            `yaml:"method"`
	URLTemplate  string            `yaml:"url_template"`
	BodyTemplate string            `yaml:"body_template"`
	Headers      map[string]string `yaml:"headers"` // Flexible headers (Authorization, Cookie, etc)
}

// Config is the main configuration structure
type Config struct {
	Request RequestConfig `yaml:"request"`
	CSV     CSVConfig     `yaml:"csv"`
	Workers int           `yaml:"workers"`
}

// AppConfig is the legacy structure for backward compatibility
// DEPRECATED: Use Config instead
type AppConfig struct {
	Token   string `yaml:"token"`
	ApiUrl  string `yaml:"url"`
	Workers int    `yaml:"workers"`
	Path    struct {
		Method   string `yaml:"method"`
		Template string `yaml:"template"`
	} `yaml:"path"`
	Payload struct {
		Template string `yaml:"template"`
	} `yaml:"payload"`
	CSV CSVConfig `yaml:"csv"`
}

// ToConfig converts legacy AppConfig to new Config structure
func (ac *AppConfig) ToConfig() *Config {
	headers := make(map[string]string)
	if ac.Token != "" {
		headers["Authorization"] = "Bearer " + ac.Token
	}
	headers["Content-Type"] = "application/json"

	return &Config{
		Request: RequestConfig{
			Method:       ac.Path.Method,
			URLTemplate:  ac.Path.Template,
			BodyTemplate: ac.Payload.Template,
			Headers:      headers,
		},
		CSV:     ac.CSV,
		Workers: ac.Workers,
	}
}

// LoadConfig reads the configuration file from the specified path and returns a Config object.
// It searches for the configuration file with the names "config.yml" or "config.yaml" in the given path.
// It supports both new and legacy config formats.
func LoadConfig(path string) (*Config, error) {
	f, err := utils.FindFiles(path, "config.yml", "config.yaml")
	if err != nil {
		return nil, fmt.Errorf("error finding config file: %w", err)
	}
	if len(f) == 0 {
		return nil, fmt.Errorf("could not find config.yml or config.yaml in %s", styles.Bold(path))
	}

	file, err := os.ReadFile(f[0])
	if err != nil {
		return nil, err
	}

	// Try new format first
	var config Config
	err = yaml.Unmarshal(file, &config)
	if err == nil && config.Request.Method != "" {
		return &config, nil
	}

	// Fallback to legacy format
	var legacyConfig AppConfig
	err = yaml.Unmarshal(file, &legacyConfig)
	if err != nil {
		return nil, err
	}

	return legacyConfig.ToConfig(), nil
}

// LegacyConfig is the legacy Config function for backward compatibility
// DEPRECATED: Use LoadConfig instead
func LegacyConfig(path string) (AppConfig, error) {
	f, err := utils.FindFiles(path, "config.yml", "config.yaml")
	if err != nil {
		return AppConfig{}, fmt.Errorf("error finding config file: %w", err)
	}
	if len(f) == 0 {
		return AppConfig{}, fmt.Errorf("could not find config.yml or config.yaml in %s", styles.Bold(path))
	}

	file, err := os.ReadFile(f[0])
	if err != nil {
		return AppConfig{}, err
	}
	var config AppConfig
	err = yaml.Unmarshal(file, &config)

	return config, err
}

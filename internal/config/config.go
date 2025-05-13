package config

import (
	"fmt"
	"os"

	"github.com/anibaldeboni/rapper/internal/styles"
	"github.com/anibaldeboni/rapper/internal/utils"
	yaml "gopkg.in/yaml.v3"
)

type CSV struct {
	Separator string   `yaml:"separator"`
	Fields    []string `yaml:"fields"`
}

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
	CSV CSV `yaml:"csv"`
}

// Config reads the configuration file from the specified path and returns an AppConfig object.
// It searches for the configuration file with the names "config.yml" or "config.yaml" in the given path.
// If the file is found, it reads and unmarshals the file content into the AppConfig object.
// If the file is not found or there is an error reading/unmarshaling the file, it returns an empty AppConfig object and an error.
func Config(path string) (AppConfig, error) {
	f, err := utils.FindFiles(path, "config.yml", "config.yaml")
	if err != nil {
		return AppConfig{}, fmt.Errorf("Error finding config file: %w", err)
	}
	if len(f) == 0 {
		return AppConfig{}, fmt.Errorf("Could not find config.yml or config.yaml in %s", styles.Bold(path))
	}

	file, err := os.ReadFile(f[0])
	if err != nil {
		return AppConfig{}, err
	}
	var config AppConfig
	err = yaml.Unmarshal(file, &config)

	return config, err
}

package config

import (
	"fmt"
	"os"

	"github.com/anibaldeboni/rapper/internal/styles"
	"github.com/anibaldeboni/rapper/internal/utils"
	yaml "gopkg.in/yaml.v3"
)

type CSV struct {
	Fields    []string `yaml:"fields"`
	Separator string   `yaml:"separator"`
}

type App struct {
	Token  string `yaml:"token"`
	ApiUrl string `yaml:"url"`
	Path   struct {
		Method   string `yaml:"method"`
		Template string `yaml:"template"`
	} `yaml:"path"`
	Payload struct {
		Template string `yaml:"template"`
	} `yaml:"payload"`
	CSV CSV `yaml:"csv"`
}

func Config(path string) (App, error) {
	f, errs := utils.FindFiles(path, "config.yml", "config.yaml")
	if len(errs) > 0 || len(f) == 0 {
		return App{}, fmt.Errorf("Could not find config.yml or config.yaml in %s", styles.Bold(path))
	}

	file, err := os.ReadFile(f[0])
	if err != nil {
		return App{}, err
	}
	var config App
	err = yaml.Unmarshal(file, &config)

	return config, err
}

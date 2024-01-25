package files

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"rapper/ui"
	"rapper/ui/list"
	"strings"
	"text/template"

	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Token  string `yaml:"token"`
	ApiUrl string `yaml:"url"`
	Path   struct {
		Method   string `yaml:"method"`
		Template string `yaml:"template"`
	} `yaml:"path"`
	Payload struct {
		Template string `yaml:"template"`
	} `yaml:"payload"`
	CSV []string `yaml:"csv"`
}

type CSVLine map[string]string
type CSV []CSVLine

func Config(path string) (AppConfig, error) {
	file, err := os.ReadFile(path + "/config.yml")
	if err != nil {
		return AppConfig{}, err
	}
	var config AppConfig
	err = yaml.Unmarshal(file, &config)

	return config, err
}

func ChooseFile(path string) (string, error) {
	files, err := findCsvFiles(path)
	if err != nil {
		return "", err
	}
	p := tea.NewProgram(list.BuildList(files, ui.Bold("Choose a CSV to process")))

	m, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("Oh no: %s", ui.Red(err.Error()))
	}

	if m, ok := m.(list.Model); ok && m.Choice != "" {
		return strings.TrimSpace(m.Choice), nil
	}
	return "", errors.New("No file selected")
}

func NewTemplate(name string, templ string) *template.Template {
	return template.Must(template.New(name).Parse(templ))
}

func RenderTemplate[T map[string]string](t *template.Template, variables T) *bytes.Buffer {
	var result string
	buf := bytes.NewBufferString(result)
	t.Execute(buf, variables)

	return buf
}

func contains[T comparable](slice []T, element T) bool {
	for _, a := range slice {
		if a == element {
			return true
		}
	}
	return false
}

func FilterCSV(csv CSV, fields []string) CSV {
	var filteredCSV CSV
	for _, line := range csv {
		element := CSVLine{}
		for key, value := range line {
			if contains(fields, key) {
				element[key] = value
			}
		}
		filteredCSV = append(filteredCSV, element)
	}
	return filteredCSV
}

func MapCSV(filePath string) (CSV, error) {
	csvfile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer csvfile.Close()

	reader := csv.NewReader(csvfile)
	rawCSV, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var header []string
	var mappedCSV CSV
	for lineNum, record := range rawCSV {
		if lineNum == 0 {
			header = record
			// for i := 0; i < len(record); i++ {
			// 	header = append(header, strings.TrimSpace(record[i]))
			// }
		} else {
			line := CSVLine{}
			for i := 0; i < len(record); i++ {
				line[header[i]] = record[i]
			}
			mappedCSV = append(mappedCSV, line)
		}
	}

	return mappedCSV, nil
}

func findCsvFiles(path string) ([]string, error) {
	pattern := "*.csv"
	files, err := filepath.Glob(path + "/" + pattern)
	if err != nil {
		return nil, err
	}
	return files, nil
}

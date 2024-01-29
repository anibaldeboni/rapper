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
	CSV struct {
		Fields    []string `yaml:"fields"`
		Separator string   `yaml:"separator"`
	} `yaml:"csv"`
}

type CSVLine map[string]string
type CSV struct {
	Name  string
	Lines []CSVLine
}

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

	if len(files) == 0 {
		return "", fmt.Errorf("No CSV files found in %s", ui.Bold(path))
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

func MapCSV(filePath string, separator string, fields []string) (CSV, error) {
	if len(separator) > 1 {
		return CSV{}, fmt.Errorf("Invalid separator: %s", ui.Bold(separator))
	}
	if separator == "" {
		separator = ","
	}
	csvfile, err := os.Open(filePath)
	if err != nil {
		return CSV{}, err
	}
	defer csvfile.Close()
	fileName := filepath.Base(filePath)
	reader := csv.NewReader(csvfile)
	reader.Comma = []rune(separator)[0]
	rawCSV, err := reader.ReadAll()
	if err != nil {
		return CSV{}, err
	}

	var header []string
	var csvLines []CSVLine
	var fieldsPosition []int
	for lineNum, record := range rawCSV {
		if lineNum == 0 {
			for i := 0; i < len(record); i++ {
				if contains(fields, record[i]) {
					fieldsPosition = append(fieldsPosition, i)
				}
				header = append(header, strings.TrimSpace(record[i]))
			}
		} else {
			line := CSVLine{}
			for i := 0; i < len(record); i++ {
				if contains(fieldsPosition, i) {
					line[header[i]] = record[i]
				}
			}
			csvLines = append(csvLines, line)
		}
	}

	return CSV{
		Name:  fileName,
		Lines: csvLines,
	}, nil
}

func IsDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}

	if fileInfo.IsDir() {
		return true
	}

	return false
}

func findCsvFiles(path string) ([]string, error) {
	pattern := "*.csv"
	files, err := filepath.Glob(path + "/" + pattern)
	if err != nil {
		return nil, err
	}
	return files, nil
}

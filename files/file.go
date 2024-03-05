package files

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"text/template"

	"github.com/anibaldeboni/rapper/cli/ui"

	yaml "gopkg.in/yaml.v3"
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
	f, errs := FindFiles(path, "config.yml", "config.yaml")
	if len(errs) > 0 || len(f) == 0 {
		return AppConfig{}, fmt.Errorf("Could not find config.yml or config.yaml in %s", ui.Bold(path))
	}

	file, err := os.ReadFile(f[0])
	if err != nil {
		return AppConfig{}, err
	}
	var config AppConfig
	err = yaml.Unmarshal(file, &config)

	return config, err
}

func NewTemplate(name string, templ string) *template.Template {
	return template.Must(template.New(name).Parse(templ))
}

func RenderTemplate[T map[string]string](t *template.Template, variables T) *bytes.Buffer {
	var result string
	buf := bytes.NewBufferString(result)
	_ = t.Execute(buf, variables)

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

	reader := csv.NewReader(csvfile)
	reader.Comma = []rune(separator)[0]
	rawCSV, err := reader.ReadAll()
	if err != nil {
		return CSV{}, err
	}

	var header = rawCSV[0]
	var lines []CSVLine
	var fieldsPosition []int
	if len(fields) == 0 {
		fields = header
	}
	for _, field := range fields {
		fieldsPosition = append(fieldsPosition, slices.Index(header, field))
	}
	for i := 1; i < len(rawCSV); i++ {
		record := rawCSV[i]
		line := CSVLine{}
		for i, r := range record {
			if contains(fieldsPosition, i) {
				line[header[i]] = r
			}
		}
		lines = append(lines, line)
	}

	return CSV{
		Name:  filepath.Base(filePath),
		Lines: lines,
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

// FindFiles returns a list of files that match the given pattern in the given directory.
func FindFiles(dir string, f ...string) ([]string, []error) {
	files := []string{}
	errs := []error{}
	for _, file := range f {
		found, err := filepath.Glob(dir + "/" + file)
		if err != nil {
			errs = append(errs, err)
		}
		files = append(files, found...)
	}

	return files, errs
}

package files

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/anibaldeboni/rapper/internal/styles"
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

func Config(path string) (AppConfig, error) {
	f, errs := FindFiles(path, "config.yml", "config.yaml")
	if len(errs) > 0 || len(f) == 0 {
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

func NewTemplate(name string, templ string) *template.Template {
	return template.Must(template.New(name).Parse(templ))
}

func RenderTemplate[T map[string]string](t *template.Template, variables T) *bytes.Buffer {
	var result string
	buf := bytes.NewBufferString(result)
	_ = t.Execute(buf, variables)

	return buf
}

// CSVLine represents a line in the CSV file.
type CSVLine map[string]string

// CSV represents the CSV file.
type CSV struct {
	Name  string
	Lines []CSVLine
}

// MapCSV reads a CSV file and maps its contents into a custom data structure.
func MapCSV(filePath string, separator string, fields []string) (CSV, error) {
	if len(separator) > 1 {
		return CSV{}, fmt.Errorf("Invalid separator: %s", separator)
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
	if len(rawCSV) == 0 {
		return CSV{}, fmt.Errorf("Empty file: %s", filePath)
	}

	header := rawCSV[0]
	lines := make([]CSVLine, 0, len(rawCSV)-1)
	fieldPositions := make(map[string]int, len(fields))
	if len(fields) == 0 {
		fields = header
	}
	for i, field := range fields {
		fieldPositions[field] = i
	}

	for _, record := range rawCSV[1:] {
		line := CSVLine{}
		for i, r := range record {
			if _, ok := fieldPositions[header[i]]; ok {
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

// FindFiles takes a directory path and a list of file patterns as input and returns a list of files that match the patterns in the given directory.
// It also returns a list of any errors encountered during the process.
func FindFiles(dir string, f ...string) ([]string, []error) {
	var files []string
	var errs []error
	for _, file := range f {
		found, err := filepath.Glob(filepath.Join(dir, file))
		if err != nil {
			errs = append(errs, err)
		}
		files = append(files, found...)
	}

	return files, errs
}

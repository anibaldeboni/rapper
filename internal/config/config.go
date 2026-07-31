package config

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

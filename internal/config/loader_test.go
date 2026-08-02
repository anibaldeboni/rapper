package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoader_Load_LegacyYAML_Rejected(t *testing.T) {
	t.Parallel()

	filePath := filepath.Join(t.TempDir(), "legacy.yml")
	legacy := "url: http://legacy\ntoken: abc\npath:\n  method: POST\n  template: /v1/x\npayload:\n  template: '{}'\ncsv:\n  fields: [id]\nworkers: 1\n"
	if err := os.WriteFile(filePath, []byte(legacy), 0o644); err != nil {
		t.Fatalf("write legacy config: %v", err)
	}

	_, err := NewLoader().Load(filePath)
	if err == nil {
		t.Fatal("Loader.Load must reject legacy YAML")
	}
	if !strings.Contains(err.Error(), "request.method is required") {
		t.Fatalf("error must name the missing new-schema field, got: %v", err)
	}
	if !strings.Contains(err.Error(), filePath) {
		t.Fatalf("error must reference the config path, got: %v", err)
	}
}

func TestLoader_Load_PartialNewSchema_Rejected(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "missing request method",
			content: "request:\n  url_template: http://example\ncsv:\n  fields: [id]\nworkers: 1\n",
			wantErr: "request.method is required",
		},
		{
			name:    "negative workers",
			content: "request:\n  method: GET\n  url_template: http://example\ncsv:\n  fields: [id]\nworkers: -1\n",
			wantErr: "workers must be >= 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			filePath := filepath.Join(t.TempDir(), "partial.yml")
			if err := os.WriteFile(filePath, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("write config: %v", err)
			}

			_, err := NewLoader().Load(filePath)
			if err == nil {
				t.Fatalf("Loader.Load must reject %s", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want it to contain %q", err, tt.wantErr)
			}
			if !strings.Contains(err.Error(), filePath) {
				t.Fatalf("error must reference the config path, got: %v", err)
			}
		})
	}
}

func TestLoader_Load_WithoutCSVFields(t *testing.T) {
	t.Parallel()

	filePath := filepath.Join(t.TempDir(), "profile.yml")
	content := `request:
  method: POST
  url_template: https://example.test/{{.id}}
csv:
  separator: ","
workers: 2
`
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := NewLoader().Load(filePath)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	if cfg.CSV.Separator != "," {
		t.Fatalf("CSV separator = %q, want %q", cfg.CSV.Separator, ",")
	}
}

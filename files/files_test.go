package files_test

import (
	"testing"

	"github.com/anibaldeboni/rapper/files"

	"github.com/stretchr/testify/assert"
)

func TestMapCSV(t *testing.T) {

	tests := []struct {
		name    string
		path    string
		fields  []string
		wantErr bool
	}{
		{
			name:    "When the file exists",
			path:    "../tests/example.csv",
			fields:  []string{"id", "street_number"},
			wantErr: false,
		},
		{
			name:    "When no field is specified",
			path:    "../tests/example.csv",
			fields:  []string{},
			wantErr: false,
		},
		{
			name:    "When the file doesn't exist",
			path:    "./no-file.csv",
			fields:  []string{"id", "street_number"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := files.MapCSV(tt.path, ",", tt.fields)

			assert.Equal(t, tt.wantErr, err != nil)

			if got.Lines != nil && len(got.Lines) > 0 {
				assert.Equal(t, len(tt.fields), len(got.Lines[0]))
			}
		})
	}
}

func TestConfig(t *testing.T) {

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "When the file exists",
			path:    "..",
			wantErr: false,
		},
		{
			name:    "When the file doesn't exist",
			path:    "./no-file.yaml",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := files.Config(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Config() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

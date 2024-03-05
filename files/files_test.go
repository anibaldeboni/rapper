package files_test

import (
	"testing"

	"github.com/anibaldeboni/rapper/files"
	"golang.org/x/exp/maps"

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
			name:    "When the file exists return the specified fields",
			path:    "../tests/example.csv",
			fields:  []string{"id", "street_number"},
			wantErr: false,
		},
		{
			name:    "When no field is specified, return all fields",
			path:    "../tests/example.csv",
			fields:  []string{"id", "street_number", "house_number", "city"},
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

			if got.Lines != nil {
				assert.ElementsMatch(t, tt.fields, maps.Keys(got.Lines[0]))
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

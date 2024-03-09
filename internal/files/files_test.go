package files_test

import (
	"os"
	"testing"

	"github.com/anibaldeboni/rapper/internal/files"
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
			path:    "../../tests/example.csv",
			fields:  []string{"id", "street_number"},
			wantErr: false,
		},
		{
			name:    "When no field is specified, return all fields",
			path:    "../../tests/example.csv",
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
				assert.ElementsMatch(t, tt.fields, maps.Keys[map[string]string](got.Lines[0]))
			}
		})
	}
}

func TestInvalidSeparator(t *testing.T) {
	// Initialize test data
	separator := ";;"
	fields := []string{"name", "age"}

	// Create a test CSV file in memory
	csvData := `name;;age;;city
                John;;25;;New York
                Jane;;30;;Los Angeles`
	tempFile, err := os.CreateTemp("", "test.csv")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	_, err = tempFile.WriteString(csvData)
	if err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}

	// Call the function under test
	_, err = files.MapCSV(tempFile.Name(), separator, fields)
	assert.Error(t, err, "Expected an error, but got nil")
}

func TestConfig(t *testing.T) {

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "When the file exists",
			path:    "../../",
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

package config_test

import (
	"testing"

	"github.com/anibaldeboni/rapper/internal/config"
)

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
			_, err := config.Config(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Config() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

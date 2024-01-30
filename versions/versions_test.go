package versions

import (
	"rapper/web"
	"rapper/web/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckForUpdate(t *testing.T) {
	client := mocks.NewHttpClient(t)
	response := web.Response{
		Status:  200,
		Body:    []byte(`[{"tag_name": "v2.0.0", "html_url": "release_url"}]`),
		Headers: *new(map[string][]string),
	}
	client.On("Get", mock.Anything, mock.Anything).Return(response, nil)

	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "When app is up-to-date",
			version: "v2.0.0",
			want:    "",
		},
		{
			name:    "When app is out-of-date",
			version: "v1.0.0",
			want:    "ℹ️  New version available: v2.0.0 (release_url)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckForUpdate(client, tt.version)
			assert.Equal(t, tt.want, got)
		})
	}
}

package versions

import (
	"io"
	"rapper/web"
	"testing"
)

type httpClientMock struct{}

func (httpClientMock) Put(url string, body io.Reader, headers map[string]string) (web.Response, error) {
	return web.Response{}, nil
}

func (httpClientMock) Post(url string, body io.Reader, headers map[string]string) (web.Response, error) {
	return web.Response{}, nil
}

func (httpClientMock) Get(url string, headers map[string]string) (web.Response, error) {
	return web.Response{
		Status:  200,
		Body:    []byte(`[{"tag_name": "v2.0.0", "html_url": "release_url"}]`),
		Headers: *new(map[string][]string),
	}, nil
}
func TestCheckForUpdate(t *testing.T) {
	client := httpClientMock{}

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
			if got != tt.want {
				t.Errorf("CheckForUpdate() = %v, want %v", got, tt.want)
			}
		})
	}
}

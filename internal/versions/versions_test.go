package versions_test

import (
	"errors"
	"testing"

	"github.com/anibaldeboni/rapper/internal/versions"
	"github.com/anibaldeboni/rapper/internal/web"
	mock_web "github.com/anibaldeboni/rapper/internal/web/mock"
	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
)

func TestCheckForUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_web.NewMockHttpClient(ctrl)
	response := web.Response{
		StatusCode: 200,
		Body:       []byte(`[{"tag_name": "v2.0.0", "html_url": "release_url"}]`),
		Headers:    *new(map[string][]string),
	}

	tests := []struct {
		name     string
		version  string
		want     string
		response web.Response
		err      error
	}{
		{
			name:     "When app is up-to-date",
			version:  "v2.0.0",
			want:     "",
			response: response,
			err:      nil,
		},
		{
			name:     "When app is out-of-date",
			version:  "v1.0.0",
			want:     "ℹ️  New version available: v2.0.0 (release_url)",
			response: response,
			err:      nil,
		},
		{
			name:     "When a request error occur",
			version:  "v1.0.0",
			want:     "",
			response: web.Response{},
			err:      errors.New("request-error"),
		},
		{
			name:    "When the request body is not a json",
			version: "v1.0.0",
			want:    "",
			response: web.Response{
				StatusCode: 200,
				Body:       []byte(`body-is-not-a-json`),
				Headers:    *new(map[string][]string),
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client.EXPECT().Get(gomock.Any(), gomock.Any()).Return(tt.response, tt.err).Times(1)
			got := versions.CheckForUpdate(client, tt.version)

			assert.Equal(t, tt.want, got)
		})
	}
}

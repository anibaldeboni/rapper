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

type want struct {
	update  bool
	version string
	url     string
}

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
		want     want
		response web.Response
		err      error
	}{
		{
			name:     "When app is up-to-date",
			version:  "v2.0.0",
			want:     want{update: false, version: "", url: ""},
			response: response,
			err:      nil,
		},
		{
			name:     "When app is out-of-date",
			version:  "v1.0.0",
			want:     want{update: true, version: "v2.0.0", url: "release_url"},
			response: response,
			err:      nil,
		},
		{
			name:     "When a request error occur",
			version:  "v1.0.0",
			want:     want{update: false, version: "", url: ""},
			response: web.Response{},
			err:      errors.New("request-error"),
		},
		{
			name:    "When the request body is not a json",
			version: "v1.0.0",
			want:    want{update: false, version: "", url: ""},
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
			subject := versions.NewUpdateChecker(client, tt.version)
			got := subject.CheckForUpdate()

			assert.Equal(t, tt.want.url, got.Url())
			assert.Equal(t, tt.want.version, got.Version())
			assert.Equal(t, tt.want.update, got.HasUpdate())
		})
	}
}

package updates_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anibaldeboni/rapper/internal/updates"

	_ "unsafe"

	"github.com/stretchr/testify/assert"
)

type want struct {
	update  bool
	version string
	url     string
}

var resp []byte

var server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(resp)
	if err != nil {
		log.Fatal(err)
	}
}))

//go:linkname releaseUrl github.com/anibaldeboni/rapper/internal/updates.releaseUrl
var releaseUrl = server.URL

func TestCheckForUpdate(t *testing.T) {
	defer server.Close()

	response := []byte(`[{"tag_name": "v2.0.0", "html_url": "release_url"}]`)

	tests := []struct {
		name     string
		version  string
		response []byte
		want     want
	}{
		{
			name:     "When app is up-to-date",
			version:  "v2.0.0",
			want:     want{update: false, version: "", url: ""},
			response: response,
		},
		{
			name:     "When app is out-of-date",
			version:  "v1.0.0",
			want:     want{update: true, version: "v2.0.0", url: "release_url"},
			response: response,
		},
		{
			name:     "When the request body is not a json",
			version:  "v1.0.0",
			want:     want{update: false, version: "", url: ""},
			response: []byte(`body-is-not-a-json`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp = tt.response
			got, ok := updates.CheckFor(tt.version)

			assert.Equal(t, tt.want.url, got.Url)
			assert.Equal(t, tt.want.version, got.Version)
			assert.Equal(t, tt.want.update, ok)
		})
	}
}

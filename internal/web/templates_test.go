package web

import (
	"testing"
	"text/template"

	edlib "github.com/hbollon/go-edlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectFields(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     []string
	}{
		{
			name:     "collects fields from url body and header templates",
			template: "https://example.test/{{.id}}/{{.id}} {{.name}} {{.region}}",
			want:     []string{"id", "name", "region"},
		},
		{
			name:     "walks nested control structures",
			template: "{{if .active}}{{range .items}}{{.id}}{{end}}{{else}}{{with .fallback}}{{.name}}{{end}}{{end}}",
			want:     []string{"active", "items", "id", "fallback", "name"},
		},
		{
			name:     "collects the base of chained fields",
			template: "{{.user.name}} {{.user.email}}",
			want:     []string{"user"},
		},
		{
			name:     "ignores literals",
			template: "plain text {{printf \"%s\" \"literal\"}}",
			want:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := template.New("test").Parse(tt.template)
			require.NoError(t, err)

			assert.Equal(t, tt.want, collectFields(tmpl))
		})
	}
}

func TestSuggestFields(t *testing.T) {
	name := "userid"
	candidates := []string{"usrid", "username", "user_id", "id"}
	want, err := edlib.FuzzySearchSet(name, candidates, 3, edlib.Levenshtein)
	require.NoError(t, err)

	assert.Equal(t, want, SuggestFields(name, candidates))
	assert.Nil(t, SuggestFields(name, nil))
}

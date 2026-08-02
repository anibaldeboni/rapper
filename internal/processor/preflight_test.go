package processor

import (
	"errors"
	"testing"

	"github.com/hbollon/go-edlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateHeaders(t *testing.T) {
	tests := []struct {
		name             string
		headers          []string
		required         []string
		wantMissing      []string
		wantSuggestions  map[string][]string
		wantErrorMessage string
	}{
		{
			name:        "reports every missing field once in lexical order",
			headers:     []string{"alpha", "betta", "gama"},
			required:    []string{"gamma", "beta", "alpha", "beta"},
			wantMissing: []string{"beta", "gamma"},
			wantSuggestions: map[string][]string{
				"beta":  {"betta", "gama", "alpha"},
				"gamma": {"gama", "alpha", "betta"},
			},
		},
		{
			name:        "allows extra columns",
			headers:     []string{"id", "name", "note"},
			required:    []string{"id", "name"},
			wantMissing: nil,
		},
		{
			name:            "empty headers have no suggestions",
			headers:         []string{},
			required:        []string{"userid"},
			wantMissing:     []string{"userid"},
			wantSuggestions: nil,
		},
		{
			name:            "blank header has no suggestions",
			headers:         []string{""},
			required:        []string{"userid"},
			wantMissing:     []string{"userid"},
			wantSuggestions: nil,
		},
		{
			name:        "no referenced fields passes",
			headers:     []string{},
			required:    []string{},
			wantMissing: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHeaders(tt.headers, tt.required)
			if len(tt.wantMissing) == 0 {
				assert.NoError(t, err)
				return
			}

			var missing *MissingFieldsError
			require.Error(t, err)
			require.True(t, errors.As(err, &missing))
			assert.Equal(t, tt.wantMissing, missing.Missing)
			assert.Equal(t, tt.wantSuggestions, missing.Suggestions)
			assert.Contains(t, missing.Error(), `missing csv fields: "`)
		})
	}
}

func TestValidateHeaders_UsesEdlibRanking(t *testing.T) {
	headers := []string{"usrid", "username", "user_id", "id"}
	err := validateHeaders(headers, []string{"userid"})

	var missing *MissingFieldsError
	require.ErrorAs(t, err, &missing)
	want, fuzzyErr := edlib.FuzzySearchSet("userid", headers, 3, edlib.Levenshtein)
	require.NoError(t, fuzzyErr)
	assert.Equal(t, want, missing.Suggestions["userid"])
}

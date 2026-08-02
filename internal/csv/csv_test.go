package csv

import (
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/anibaldeboni/rapper/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeparator(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.CSVConfig
		want rune
	}{
		{name: "default", want: ','},
		{name: "blank after trimming", cfg: config.CSVConfig{Separator: "   "}, want: ','},
		{name: "custom separator", cfg: config.CSVConfig{Separator: " ; "}, want: ';'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Separator(tt.cfg))
		})
	}
}

func TestOpen(t *testing.T) {
	_, _, err := Open(filepath.Join(t.TempDir(), "missing.csv"), ',')

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error opening file")
}

func TestHeaders(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
		wantErr string
	}{
		{name: "returns the first record", content: "id,name\n1,Ada\n", want: []string{"id", "name"}},
		{name: "reports an empty file", wantErr: "no records found in the file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.CreateTemp(t.TempDir(), "headers-*.csv")
			require.NoError(t, err)
			_, err = file.WriteString(tt.content)
			require.NoError(t, err)
			require.NoError(t, file.Close())

			reader, opened, err := Open(file.Name(), ',')
			require.NoError(t, err)
			defer opened.Close()

			got, err := Headers(reader)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReadPreview(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		separator  rune
		maxRows    int
		wantRows   int
		wantHeader []string
		wantFirst  []string
		wantErr    string
	}{
		{
			name:       "bounds large file",
			content:    "id,name\n1,Ada\n2,Bea\n3,Cia\n4,Dia\n5,Eva\n6,Fay\n7,Gia\n8,Hia\n9,Ian\n10,Jay\n11,Kai\n",
			separator:  ',',
			maxRows:    10,
			wantRows:   10,
			wantHeader: []string{"id", "name"},
			wantFirst:  []string{"1", "Ada"},
		},
		{
			name:       "returns all short files",
			content:    "id;name\n1;Ada\n2;Bea\n3;Cia\n",
			separator:  ';',
			maxRows:    10,
			wantRows:   3,
			wantHeader: []string{"id", "name"},
			wantFirst:  []string{"1", "Ada"},
		},
		{
			name:      "wraps malformed rows",
			content:   "id,name\n1\n",
			separator: ',',
			maxRows:   10,
			wantErr:   "error reading rows",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "preview.csv")
			require.NoError(t, os.WriteFile(path, []byte(tt.content), 0o600))

			got, err := ReadPreview(path, tt.separator, tt.maxRows)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantHeader, got.Headers)
			assert.Len(t, got.Rows, tt.wantRows)
			assert.Equal(t, tt.wantFirst, got.Rows[0])
		})
	}
}

func TestReadPreview_ClosesFileOnError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "malformed.csv")
	require.NoError(t, os.WriteFile(path, []byte("id,name\n1\n"), 0o600))

	_, err := ReadPreview(path, ',', 10)

	var parseErr *csv.ParseError
	require.ErrorAs(t, err, &parseErr)
	assert.NotNil(t, parseErr)

	file, openErr := os.Open(path)
	require.NoError(t, openErr)
	assert.NoError(t, file.Close())
}

func TestReadPreview_OpenError(t *testing.T) {
	_, err := ReadPreview(filepath.Join(t.TempDir(), "missing.csv"), ',', 10)

	require.Error(t, err)
	assert.True(t, errors.Is(err, os.ErrNotExist))
}

package files

import "testing"

func TestMapCSV(t *testing.T) {

	tests := []struct {
		name    string
		path    string
		fields  []string
		wantErr bool
	}{
		{
			name:    "When the file exists",
			path:    "../testdata/example.csv",
			fields:  []string{"id", "street_number"},
			wantErr: false,
		},
		{
			name:    "When no field is specified",
			path:    "../testdata/example.csv",
			fields:  []string{},
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
			got, err := MapCSV(tt.path, ",", tt.fields)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapCSV() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got.Lines != nil && len(got.Lines) > 0 {
				for k := range got.Lines[0] {
					if !contains(tt.fields, k) {
						t.Errorf("MapCSV() got = %v, want %v", k, tt.fields)
					}
				}
			}
		})
	}
}

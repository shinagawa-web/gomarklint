package testutil

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestGetTestFilePath(t *testing.T) {
	tests := []struct {
		name     string
		rel      string
		wantBase string
	}{
		{
			name:     "testdata directory",
			rel:      "testdata",
			wantBase: "testdata",
		},
		{
			name:     "testdata file",
			rel:      "testdata/sample.md",
			wantBase: "sample.md",
		},
		{
			name:     "root file",
			rel:      "README.md",
			wantBase: "README.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTestFilePath(tt.rel)

			if !strings.HasSuffix(got, tt.rel) {
				t.Errorf("GetTestFilePath(%q) = %q, should end with %q", tt.rel, got, tt.rel)
			}

			if filepath.Base(got) != tt.wantBase {
				t.Errorf("GetTestFilePath(%q) base = %q, want %q", tt.rel, filepath.Base(got), tt.wantBase)
			}

			if !filepath.IsAbs(got) {
				t.Errorf("GetTestFilePath(%q) = %q, should return absolute path", tt.rel, got)
			}
		})
	}
}

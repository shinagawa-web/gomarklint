package rule

import (
	"os"
	"testing"
)

func TestCheckHeadingLevels(t *testing.T) {
	tests := []struct {
		name     string
		filepath string
		minLevel int
		wantErr  bool
	}{
		{
			name:     "valid heading levels",
			filepath: "testdata/heading_level/valid.md",
			minLevel: 2,
			wantErr:  false,
		},
		{
			name:     "invalid jump in heading levels",
			filepath: "testdata/heading_level/invalid_jump.md",
			minLevel: 2,
			wantErr:  true,
		},
		{
			name:     "invalid first heading level",
			filepath: "testdata/heading_level/invalid_first.md",
			minLevel: 2,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := getTestFilePath(tt.filepath)
			t.Logf("Testing file: %s", path)

			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read test file: %v", err)
			}

			errors := CheckHeadingLevels(tt.filepath, string(data), tt.minLevel)
			if (len(errors) > 0) != tt.wantErr {
				t.Errorf("expected error: %v, got: %v", tt.wantErr, errors)
			}
		})
	}
}

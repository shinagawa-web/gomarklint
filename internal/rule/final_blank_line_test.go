package rule

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func getTestFilePath(rel string) string {
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "../..")
	return filepath.Join(projectRoot, rel)
}

func TestCheckFinalBlankLine(t *testing.T) {
	tests := []struct {
		name     string
		filepath string
		wantErr  bool
	}{
		{"with_blank", "testdata/final_blank/with_blank.md", false},
		{"no_blank", "testdata/final_blank/no_blank.md", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := getTestFilePath(tt.filepath)
			println("Testing file:", path)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read test file: %v", err)
			}
			errors := CheckFinalBlankLine(path, string(data))
			if (len(errors) > 0) != tt.wantErr {
				t.Errorf("expected error: %v, got %v", tt.wantErr, errors)
			}
		})
	}
}

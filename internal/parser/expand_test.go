package parser

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func setupTestFiles(t *testing.T) string {
	base := t.TempDir()

	// Directory structure:
	// base/
	//   ├── file1.md
	//   ├── file2.txt
	//   └── subdir/
	//         └── nested.md

	mustWrite := func(relPath, content string) {
		full := filepath.Join(base, relPath)
		os.MkdirAll(filepath.Dir(full), 0755)
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}
	}

	mustWrite("file1.md", "# Hello")
	mustWrite("file2.txt", "text")
	mustWrite("subdir/nested.md", "# Nested")
	mustWrite(".hidden/secret.md", "# Hidden")

	return base
}

func TestExpandPaths(t *testing.T) {
	base := setupTestFiles(t)

	tests := []struct {
		name     string
		input    []string
		wantEnds []string
	}{
		{
			name:     "single file",
			input:    []string{filepath.Join(base, "file1.md")},
			wantEnds: []string{"file1.md"},
		},
		{
			name:     "directory with nested md",
			input:    []string{base},
			wantEnds: []string{"file1.md", "subdir/nested.md"},
		},
		{
			name:     "non-md file",
			input:    []string{filepath.Join(base, "file2.txt")},
			wantEnds: []string{},
		},
		{
			name:     "nonexistent path",
			input:    []string{filepath.Join(base, "nonexistent.md")},
			wantEnds: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandPaths(tt.input, []string{})
			if err != nil {
				t.Fatalf("ExpandPaths failed: %v", err)
			}

			var gotEnds []string
			for _, path := range got {
				gotEnds = append(gotEnds, filepath.ToSlash(path[len(base)+1:]))
			}

			if !reflect.DeepEqual(sorted(gotEnds), sorted(tt.wantEnds)) {
				t.Errorf("expected %v, got %v", tt.wantEnds, gotEnds)
			}
		})
	}
	t.Run("unreadable directory", func(t *testing.T) {
		base := t.TempDir()
		badDir := filepath.Join(base, "secret")

		if err := os.Mkdir(badDir, 0755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
		if err := os.Chmod(badDir, 0000); err != nil {
			t.Skipf("cannot make directory unreadable, skipping test: %v", err)
		}
		defer os.Chmod(badDir, 0755) // cleanup

		_, err := ExpandPaths([]string{base}, []string{})
		if err != nil {
			t.Fatalf("ExpandPaths failed: %v", err)
		}
	})
}

func sorted(s []string) []string {
	clone := make([]string, len(s))
	copy(clone, s)
	sort.Strings(clone)
	return clone
}

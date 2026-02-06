package file

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

func TestReadFile(t *testing.T) {
	t.Run("successfully reads file content", func(t *testing.T) {
		path := getTestFilePath("testdata/sample.md")
		expected := "# Hello, World!\n"

		err := os.WriteFile(path, []byte(expected), 0644)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		got, err := ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile returned unexpected error: %v", err)
		}
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})

	t.Run("returns error for nonexistent file", func(t *testing.T) {
		path := getTestFilePath("testdata/does_not_exist.md")
		_, err := ReadFile(path)
		if err == nil {
			t.Error("expected error for nonexistent file, got nil")
		}
	})
}

func TestStripFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantBody string
		wantSkip int
	}{
		{
			name: "with frontmatter",
			input: `---
title: "Test"
date: 2025-01-01
---

# Hello`,
			wantBody: `# Hello`,
			wantSkip: 5,
		},
		{
			name:     "no frontmatter",
			input:    `# Hello`,
			wantBody: `# Hello`,
			wantSkip: 0,
		},
		{
			name: "incomplete frontmatter",
			input: `---
title: "Oops"`,
			wantBody: `---
title: "Oops"`,
			wantSkip: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, skip := StripFrontmatter(tt.input)
			if body != tt.wantBody {
				t.Errorf("got body %q, want %q", body, tt.wantBody)
			}
			if skip != tt.wantSkip {
				t.Errorf("got skip %d, want %d", skip, tt.wantSkip)
			}
		})
	}
}

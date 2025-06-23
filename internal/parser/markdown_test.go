package parser

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// ヘルパー関数：プロジェクトルートからの相対パスを解決
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

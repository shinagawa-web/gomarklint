package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// captureStdout redirects os.Stdout for the duration of f and returns what was written.
func captureStdout(f func() error) (string, error) {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	err := f()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String(), err
}

func TestExecute(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "valid.md")
	os.WriteFile(f, []byte("## Hello\n\nWorld.\n"), 0644)

	rootCmd.SetArgs([]string{f, "--config", "/nonexistent/.gomarklint.json"})
	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	_, err := captureStdout(func() error {
		return Execute()
	})
	if err != nil {
		t.Errorf("expected no error from Execute, got: %v", err)
	}
}

func TestInitCmd_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	original, _ := os.Getwd()
	os.Chdir(dir)
	t.Cleanup(func() { os.Chdir(original) })

	rootCmd.SetArgs([]string{"init"})
	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	_, err := captureStdout(func() error {
		return Execute()
	})
	if err != nil {
		t.Errorf("expected no error from init, got: %v", err)
	}
	if _, statErr := os.Stat(".gomarklint.json"); statErr != nil {
		t.Error("expected .gomarklint.json to be created")
	}
}

func TestExecute_WithOutputAndSeverityFlags(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "valid.md")
	os.WriteFile(f, []byte("## Hello\n\nWorld.\n"), 0644)

	rootCmd.SetArgs([]string{f, "--config", "/nonexistent/.gomarklint.json", "--output", "json", "--severity", "warning"})
	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	out, err := captureStdout(func() error {
		return Execute()
	})
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if !strings.Contains(out, `"files"`) {
		t.Errorf("expected JSON output, got: %s", out)
	}
}

func TestInitCmd_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	original, _ := os.Getwd()
	os.Chdir(dir)
	t.Cleanup(func() { os.Chdir(original) })

	os.WriteFile(".gomarklint.json", []byte("{}"), 0644)

	rootCmd.SetArgs([]string{"init"})
	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	_, err := captureStdout(func() error {
		return Execute()
	})
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

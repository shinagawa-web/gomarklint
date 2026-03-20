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
func captureStdout(t *testing.T, f func() error) (string, error) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	runErr := f()
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()
	return buf.String(), runErr
}

func TestExecute(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "valid.md")
	if err := os.WriteFile(f, []byte("## Hello\n\nWorld.\n"), 0644); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetArgs([]string{f, "--config", "/nonexistent/.gomarklint.json"})
	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	_, err := captureStdout(t, func() error {
		return Execute()
	})
	if err != nil {
		t.Errorf("expected no error from Execute, got: %v", err)
	}
}

func TestInitCmd_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	original, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(original) })

	rootCmd.SetArgs([]string{"init"})
	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	_, err := captureStdout(t, func() error {
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
	if err := os.WriteFile(f, []byte("## Hello\n\nWorld.\n"), 0644); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetArgs([]string{f, "--config", "/nonexistent/.gomarklint.json", "--output", "json", "--severity", "warning"})
	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	out, err := captureStdout(t, func() error {
		return Execute()
	})
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if !strings.Contains(out, `"files"`) {
		t.Errorf("expected JSON output, got: %s", out)
	}
}

func TestInitCmd_WriteError(t *testing.T) {
	dir := t.TempDir()
	// Make the directory unwritable so os.WriteFile fails.
	if err := os.Chmod(dir, 0555); err != nil {
		t.Skip("cannot set directory permissions:", err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0755) })

	original, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(original) })

	rootCmd.SetArgs([]string{"init"})
	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	_, err := captureStdout(t, func() error {
		return Execute()
	})
	if err == nil || !strings.Contains(err.Error(), "failed to write config file") {
		t.Errorf("expected 'failed to write config file' error, got: %v", err)
	}
}

func TestInitCmd_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	original, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(original) })

	if err := os.WriteFile(".gomarklint.json", []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetArgs([]string{"init"})
	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	_, err := captureStdout(t, func() error {
		return Execute()
	})
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

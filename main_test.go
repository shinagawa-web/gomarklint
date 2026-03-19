package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMain_HappyPath(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "valid.md")
	os.WriteFile(f, []byte("## Hello\n\nWorld.\n"), 0644)

	old := os.Args
	os.Args = []string{"gomarklint", f, "--config", "/nonexistent/.gomarklint.json"}
	defer func() { os.Args = old }()

	// Capture stdout to avoid test noise
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w
	defer func() {
		w.Close()
		os.Stdout = oldStdout
		r.Close()
	}()

	main() // should return normally without calling osExit
}

func TestMain_LintViolation(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "invalid.md")
	os.WriteFile(f, []byte("# H1 heading\n"), 0644)

	old := os.Args
	os.Args = []string{"gomarklint", f, "--config", "/nonexistent/.gomarklint.json"}
	defer func() { os.Args = old }()

	exitCode := -1
	oldExit := osExit
	osExit = func(code int) { exitCode = code }
	defer func() { osExit = oldExit }()

	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w
	defer func() {
		w.Close()
		os.Stdout = oldStdout
		r.Close()
	}()

	main()
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}

func TestMain_GenericError(t *testing.T) {
	dir := t.TempDir()
	badConfig := filepath.Join(dir, "bad.json")
	os.WriteFile(badConfig, []byte("{invalid json}"), 0644)

	old := os.Args
	os.Args = []string{"gomarklint", "--config", badConfig, "somefile.md"}
	defer func() { os.Args = old }()

	exitCode := -1
	oldExit := osExit
	osExit = func(code int) { exitCode = code }
	defer func() { osExit = oldExit }()

	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w
	oldStderr := os.Stderr
	os.Stderr = w
	defer func() {
		w.Close()
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		r.Close()
	}()

	main()
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}

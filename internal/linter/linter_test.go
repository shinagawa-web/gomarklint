package linter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shinagawa-web/gomarklint/v2/internal/config"
)

// off returns a disabled RuleConfig.
func off() *config.RuleConfig {
	return &config.RuleConfig{Enabled: false, Severity: config.SeverityOff, Options: map[string]interface{}{}}
}

// on returns an enabled RuleConfig with error severity.
func on() *config.RuleConfig {
	return &config.RuleConfig{Enabled: true, Severity: config.SeverityError, Options: map[string]interface{}{}}
}

// allOff returns a default config with every rule disabled.
func allOff() config.Config {
	cfg := config.Default()
	for k := range cfg.Rules {
		cfg.Rules[k] = off()
	}
	return cfg
}

func TestNew(t *testing.T) {
	cfg := config.Default()
	cfg.Rules["external-link"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options: map[string]interface{}{
			"skipPatterns":   []interface{}{"https://example\\.com/.*"},
			"timeoutSeconds": float64(5),
		},
	}

	linter := New(cfg)
	if len(linter.compiledPatterns) != 1 {
		t.Errorf("expected 1 compiled pattern, got %d", len(linter.compiledPatterns))
	}
}

func TestNew_InvalidPattern(t *testing.T) {
	cfg := config.Default()
	cfg.Rules["external-link"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options: map[string]interface{}{
			"skipPatterns":   []interface{}{"[invalid("},
			"timeoutSeconds": float64(5),
		},
	}

	linter := New(cfg)
	if len(linter.compiledPatterns) != 0 {
		t.Errorf("expected 0 compiled patterns (invalid pattern should be skipped), got %d", len(linter.compiledPatterns))
	}
}

func TestRun_NoErrors(t *testing.T) {
	cfg := allOff()

	linter := New(cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(testFile, []byte("# Hello\n\nThis is a test.\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := linter.Run([]string{testFile})

	if result.TotalErrors != 0 {
		t.Errorf("expected 0 errors, got %d", result.TotalErrors)
	}
	if result.TotalLines != 4 {
		t.Errorf("expected 4 lines, got %d", result.TotalLines)
	}
	if len(result.OrderedPaths) != 1 {
		t.Errorf("expected 1 file path, got %d", len(result.OrderedPaths))
	}
}

func TestRun_WithErrors(t *testing.T) {
	cfg := allOff()
	cfg.Rules["heading-level"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options:  map[string]interface{}{"minLevel": float64(2)},
	}

	linter := New(cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(testFile, []byte("# Title\n\nContent\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := linter.Run([]string{testFile})

	if result.TotalErrors == 0 {
		t.Error("expected errors, got 0")
	}
	if len(result.Errors[testFile]) == 0 {
		t.Error("expected errors for test file")
	}
}

func TestRun_MultipleFiles(t *testing.T) {
	cfg := allOff()

	linter := New(cfg)

	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.md")
	file2 := filepath.Join(tmpDir, "file2.md")
	file3 := filepath.Join(tmpDir, "file3.md")

	for _, f := range []string{file1, file2, file3} {
		if err := os.WriteFile(f, []byte("# Test\n\nContent\n"), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}
	}

	result := linter.Run([]string{file1, file2, file3})

	if len(result.OrderedPaths) != 3 {
		t.Errorf("expected 3 files, got %d", len(result.OrderedPaths))
	}
	for i := 0; i < len(result.OrderedPaths)-1; i++ {
		if result.OrderedPaths[i] > result.OrderedPaths[i+1] {
			t.Errorf("paths are not sorted: %s > %s", result.OrderedPaths[i], result.OrderedPaths[i+1])
		}
	}
}

func TestRun_UnclosedCodeBlock(t *testing.T) {
	cfg := allOff()
	cfg.Rules["unclosed-code-block"] = on()

	linter := New(cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "unclosed.md")
	if err := os.WriteFile(testFile, []byte("# Test\n\n```go\ncode here\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := linter.Run([]string{testFile})

	if result.TotalErrors == 0 {
		t.Error("expected error for unclosed code block")
	}
}

func TestRun_EmptyAltText(t *testing.T) {
	cfg := allOff()
	cfg.Rules["empty-alt-text"] = on()

	linter := New(cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty-alt.md")
	if err := os.WriteFile(testFile, []byte("# Test\n\n![](image.png)\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := linter.Run([]string{testFile})

	if result.TotalErrors == 0 {
		t.Error("expected error for empty alt text")
	}
}

func TestRun_FileReadError(t *testing.T) {
	cfg := config.Default()
	linter := New(cfg)

	result := linter.Run([]string{"/non/existent/file.md"})

	if len(result.FailedFiles) != 1 {
		t.Errorf("expected 1 failed file, got %d", len(result.FailedFiles))
	}
	if _, exists := result.FailedFiles["/non/existent/file.md"]; !exists {
		t.Error("expected /non/existent/file.md in FailedFiles")
	}
	if len(result.Errors) != 0 {
		t.Errorf("expected no results for non-existent file, got %d", len(result.Errors))
	}
}

func TestRun_DuplicateHeadings(t *testing.T) {
	cfg := allOff()
	cfg.Rules["duplicate-heading"] = on()

	linter := New(cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "duplicate.md")
	if err := os.WriteFile(testFile, []byte("# Title\n\n## Section\n\n## Section\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := linter.Run([]string{testFile})

	if result.TotalErrors == 0 {
		t.Error("expected error for duplicate headings")
	}
}

func TestRun_NoMultipleBlankLines(t *testing.T) {
	cfg := allOff()
	cfg.Rules["no-multiple-blank-lines"] = on()

	linter := New(cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "blank.md")
	if err := os.WriteFile(testFile, []byte("# Title\n\n\n\nContent\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := linter.Run([]string{testFile})

	if result.TotalErrors == 0 {
		t.Error("expected error for multiple blank lines")
	}
}

func TestRun_NoSetextHeadings(t *testing.T) {
	cfg := allOff()
	cfg.Rules["no-setext-headings"] = on()

	linter := New(cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "setext.md")
	if err := os.WriteFile(testFile, []byte("Title\n=====\n\nContent\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := linter.Run([]string{testFile})

	if result.TotalErrors == 0 {
		t.Error("expected error for setext headings")
	}
}

func TestRun_FinalBlankLine(t *testing.T) {
	cfg := allOff()
	cfg.Rules["final-blank-line"] = on()

	linter := New(cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "nofinal.md")
	if err := os.WriteFile(testFile, []byte("# Title\n\nContent"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := linter.Run([]string{testFile})

	if result.TotalErrors == 0 {
		t.Error("expected error for missing final blank line")
	}
}

func TestRun_LinkCheck(t *testing.T) {
	cfg := allOff()
	cfg.Rules["external-link"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options: map[string]interface{}{
			"timeoutSeconds": float64(5),
			"skipPatterns":   []interface{}{},
		},
	}

	linter := New(cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "links.md")
	content := "# Test\n\n[Valid Link](https://example.com)\n\n[Another Link](https://www.ietf.org/rfc/rfc2606.txt)\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := linter.Run([]string{testFile})

	if result.TotalLinksChecked != 2 {
		t.Errorf("expected 2 links checked, got %d", result.TotalLinksChecked)
	}
}

func TestRun_LinkCheckWithSkipPattern(t *testing.T) {
	cfg := allOff()
	cfg.Rules["external-link"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options: map[string]interface{}{
			"timeoutSeconds": float64(5),
			"skipPatterns":   []interface{}{"https://example\\.com/.*"},
		},
	}

	linter := New(cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "links.md")
	content := "# Test\n\n[Skipped](https://example.com/skip)\n\n[Checked](https://www.ietf.org/rfc/rfc2606.txt)\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := linter.Run([]string{testFile})

	if result.TotalLinksChecked != 1 {
		t.Errorf("expected 1 link checked (skipped link not counted), got %d", result.TotalLinksChecked)
	}
}

func TestRun_DuplicatePaths(t *testing.T) {
	cfg := allOff()

	linter := New(cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(testFile, []byte("# Test\n\nContent\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := linter.Run([]string{testFile, testFile, testFile})

	if len(result.OrderedPaths) != 1 {
		t.Errorf("expected 1 unique path, got %d", len(result.OrderedPaths))
	}
	if result.TotalLines != 4 {
		t.Errorf("expected 4 lines (counted once), got %d", result.TotalLines)
	}
}

func TestLintContent_NoErrors(t *testing.T) {
	cfg := allOff()

	linter := New(cfg)

	errors, lineCount, linksChecked := linter.LintContent("test.md", "# Hello\n\nThis is a test.\n")

	if len(errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errors))
	}
	if lineCount != 4 {
		t.Errorf("expected 4 lines, got %d", lineCount)
	}
	if linksChecked != 0 {
		t.Errorf("expected 0 links checked, got %d", linksChecked)
	}
}

func TestLintContent_WithErrors(t *testing.T) {
	cfg := allOff()
	cfg.Rules["heading-level"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options:  map[string]interface{}{"minLevel": float64(2)},
	}

	linter := New(cfg)

	errors, lineCount, _ := linter.LintContent("test.md", "# Title\n\nContent\n")

	if len(errors) == 0 {
		t.Error("expected at least 1 error (heading level)")
	}
	if lineCount != 4 {
		t.Errorf("expected 4 lines, got %d", lineCount)
	}
}

func TestRun_WarningSeverity(t *testing.T) {
	cfg := allOff()
	cfg.Rules["no-setext-headings"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityWarning,
		Options:  map[string]interface{}{},
	}

	linter := New(cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "setext.md")
	content := "## Intro\n\nSection One\n===========\n\nContent\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := linter.Run([]string{testFile})

	if result.TotalWarnings == 0 {
		t.Error("expected TotalWarnings > 0 for warning-severity rule violation")
	}
	if result.TotalErrors != 0 {
		t.Errorf("expected TotalErrors=0, got %d", result.TotalErrors)
	}

	// Each violation should be tagged with severity="warning"
	for _, errs := range result.Errors {
		for _, e := range errs {
			if e.Severity != "warning" {
				t.Errorf("expected severity=warning, got %q", e.Severity)
			}
		}
	}
}

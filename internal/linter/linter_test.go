package linter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shinagawa-web/gomarklint/internal/config"
)

func TestNew(t *testing.T) {
	cfg := config.Default()
	cfg.SkipLinkPatterns = []string{"https://example\\.com/.*"}
	cfg.EnableLinkCheck = true

	linter, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if linter.config.EnableLinkCheck != true {
		t.Error("expected EnableLinkCheck to be true")
	}
	if len(linter.compiledPatterns) != 1 {
		t.Errorf("expected 1 compiled pattern, got %d", len(linter.compiledPatterns))
	}
}

func TestNew_InvalidPattern(t *testing.T) {
	cfg := config.Default()
	cfg.SkipLinkPatterns = []string{"[invalid("}
	cfg.EnableLinkCheck = true

	linter, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(linter.compiledPatterns) != 0 {
		t.Errorf("expected 0 compiled patterns (invalid pattern should be skipped), got %d", len(linter.compiledPatterns))
	}
}

func TestRun_NoErrors(t *testing.T) {
	cfg := config.Default()
	cfg.EnableHeadingLevelCheck = false
	cfg.EnableDuplicateHeadingCheck = false
	cfg.EnableNoMultipleBlankLinesCheck = false
	cfg.EnableFinalBlankLineCheck = false
	cfg.EnableNoSetextHeadingsCheck = false

	linter, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	content := "# Hello\n\nThis is a test.\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := linter.Run([]string{testFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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
	cfg := config.Default()
	cfg.EnableHeadingLevelCheck = true
	cfg.MinHeadingLevel = 2
	cfg.EnableDuplicateHeadingCheck = false
	cfg.EnableNoMultipleBlankLinesCheck = false
	cfg.EnableFinalBlankLineCheck = false
	cfg.EnableNoSetextHeadingsCheck = false

	linter, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	content := "# Title\n\nContent\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := linter.Run([]string{testFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalErrors == 0 {
		t.Error("expected errors, got 0")
	}
	if len(result.Errors[testFile]) == 0 {
		t.Error("expected errors for test file")
	}
}

func TestRun_MultipleFiles(t *testing.T) {
	cfg := config.Default()
	cfg.EnableHeadingLevelCheck = false
	cfg.EnableDuplicateHeadingCheck = false
	cfg.EnableNoMultipleBlankLinesCheck = false
	cfg.EnableFinalBlankLineCheck = false
	cfg.EnableNoSetextHeadingsCheck = false

	linter, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.md")
	file2 := filepath.Join(tmpDir, "file2.md")
	file3 := filepath.Join(tmpDir, "file3.md")

	for _, f := range []string{file1, file2, file3} {
		if err := os.WriteFile(f, []byte("# Test\n\nContent\n"), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}
	}

	result, err := linter.Run([]string{file1, file2, file3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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
	cfg := config.Default()
	cfg.EnableHeadingLevelCheck = false
	cfg.EnableDuplicateHeadingCheck = false
	cfg.EnableNoMultipleBlankLinesCheck = false
	cfg.EnableFinalBlankLineCheck = false
	cfg.EnableNoSetextHeadingsCheck = false

	linter, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "unclosed.md")
	content := "# Test\n\n```go\ncode here\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := linter.Run([]string{testFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalErrors == 0 {
		t.Error("expected error for unclosed code block")
	}

	if len(result.Errors[testFile]) == 0 {
		t.Error("expected at least one error for unclosed code block")
	}
}

func TestRun_EmptyAltText(t *testing.T) {
	cfg := config.Default()
	cfg.EnableHeadingLevelCheck = false
	cfg.EnableDuplicateHeadingCheck = false
	cfg.EnableNoMultipleBlankLinesCheck = false
	cfg.EnableFinalBlankLineCheck = false
	cfg.EnableNoSetextHeadingsCheck = false

	linter, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty-alt.md")
	content := "# Test\n\n![](image.png)\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := linter.Run([]string{testFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalErrors == 0 {
		t.Error("expected error for empty alt text")
	}

	if len(result.Errors[testFile]) == 0 {
		t.Error("expected at least one error for empty alt text")
	}
}

func TestRun_FileReadError(t *testing.T) {
	cfg := config.Default()
	linter, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := linter.Run([]string{"/non/existent/file.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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
	cfg := config.Default()
	cfg.EnableHeadingLevelCheck = false
	cfg.EnableDuplicateHeadingCheck = true
	cfg.EnableNoMultipleBlankLinesCheck = false
	cfg.EnableFinalBlankLineCheck = false
	cfg.EnableNoSetextHeadingsCheck = false

	linter, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "duplicate.md")
	content := "# Title\n\n## Section\n\n## Section\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := linter.Run([]string{testFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalErrors == 0 {
		t.Error("expected error for duplicate headings")
	}
}

func TestRun_NoMultipleBlankLines(t *testing.T) {
	cfg := config.Default()
	cfg.EnableHeadingLevelCheck = false
	cfg.EnableDuplicateHeadingCheck = false
	cfg.EnableNoMultipleBlankLinesCheck = true
	cfg.EnableFinalBlankLineCheck = false
	cfg.EnableNoSetextHeadingsCheck = false

	linter, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "blank.md")
	content := "# Title\n\n\n\nContent\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := linter.Run([]string{testFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalErrors == 0 {
		t.Error("expected error for multiple blank lines")
	}
}

func TestRun_NoSetextHeadings(t *testing.T) {
	cfg := config.Default()
	cfg.EnableHeadingLevelCheck = false
	cfg.EnableDuplicateHeadingCheck = false
	cfg.EnableNoMultipleBlankLinesCheck = false
	cfg.EnableFinalBlankLineCheck = false
	cfg.EnableNoSetextHeadingsCheck = true

	linter, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "setext.md")
	content := "Title\n=====\n\nContent\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := linter.Run([]string{testFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalErrors == 0 {
		t.Error("expected error for setext headings")
	}
}

func TestRun_FinalBlankLine(t *testing.T) {
	cfg := config.Default()
	cfg.EnableHeadingLevelCheck = false
	cfg.EnableDuplicateHeadingCheck = false
	cfg.EnableNoMultipleBlankLinesCheck = false
	cfg.EnableFinalBlankLineCheck = true
	cfg.EnableNoSetextHeadingsCheck = false

	linter, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "nofinal.md")
	content := "# Title\n\nContent"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := linter.Run([]string{testFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalErrors == 0 {
		t.Error("expected error for missing final blank line")
	}
}

func TestRun_LinkCheck(t *testing.T) {
	cfg := config.Default()
	cfg.EnableHeadingLevelCheck = false
	cfg.EnableDuplicateHeadingCheck = false
	cfg.EnableNoMultipleBlankLinesCheck = false
	cfg.EnableFinalBlankLineCheck = false
	cfg.EnableNoSetextHeadingsCheck = false
	cfg.EnableLinkCheck = true
	cfg.LinkCheckTimeoutSeconds = 5

	linter, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "links.md")
	content := "# Test\n\n[Valid Link](https://example.com)\n\n[Another Link](https://www.ietf.org/rfc/rfc2606.txt)\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := linter.Run([]string{testFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalLinksChecked != 2 {
		t.Errorf("expected 2 links checked, got %d", result.TotalLinksChecked)
	}
}

func TestRun_LinkCheckWithSkipPattern(t *testing.T) {
	cfg := config.Default()
	cfg.EnableHeadingLevelCheck = false
	cfg.EnableDuplicateHeadingCheck = false
	cfg.EnableNoMultipleBlankLinesCheck = false
	cfg.EnableFinalBlankLineCheck = false
	cfg.EnableNoSetextHeadingsCheck = false
	cfg.EnableLinkCheck = true
	cfg.SkipLinkPatterns = []string{"https://example\\.com/.*"}
	cfg.LinkCheckTimeoutSeconds = 5

	linter, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "links.md")
	content := "# Test\n\n[Skipped](https://example.com/skip)\n\n[Checked](https://www.ietf.org/rfc/rfc2606.txt)\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := linter.Run([]string{testFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalLinksChecked != 1 {
		t.Errorf("expected 1 link checked (skipped link not counted), got %d", result.TotalLinksChecked)
	}
}

func TestRun_DuplicatePaths(t *testing.T) {
	cfg := config.Default()
	cfg.EnableHeadingLevelCheck = false
	cfg.EnableDuplicateHeadingCheck = false
	cfg.EnableNoMultipleBlankLinesCheck = false
	cfg.EnableFinalBlankLineCheck = false
	cfg.EnableNoSetextHeadingsCheck = false

	linter, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	content := "# Test\n\nContent\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Pass the same file multiple times
	result, err := linter.Run([]string{testFile, testFile, testFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only process once
	if len(result.OrderedPaths) != 1 {
		t.Errorf("expected 1 unique path, got %d", len(result.OrderedPaths))
	}

	// Should count lines only once
	expectedLines := 4
	if result.TotalLines != expectedLines {
		t.Errorf("expected %d lines (counted once), got %d", expectedLines, result.TotalLines)
	}
}

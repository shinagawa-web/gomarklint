package linter

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/shinagawa-web/gomarklint/v2/internal/config"
)

// mustNew calls New and fails the test if it returns an error.
func mustNew(t *testing.T, cfg config.Config) *Linter {
	t.Helper()
	l, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return l
}

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

	linter := mustNew(t, cfg)
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

	linter := mustNew(t, cfg)
	if len(linter.compiledPatterns) != 0 {
		t.Errorf("expected 0 compiled patterns (invalid pattern should be skipped), got %d", len(linter.compiledPatterns))
	}
}

func TestRun_NoErrors(t *testing.T) {
	cfg := allOff()

	linter := mustNew(t, cfg)

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

	linter := mustNew(t, cfg)

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

	linter := mustNew(t, cfg)

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

	linter := mustNew(t, cfg)

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

func TestRun_FencedCodeLanguage(t *testing.T) {
	cfg := allOff()
	cfg.Rules["fenced-code-language"] = on()

	linter := mustNew(t, cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "no-lang.md")
	if err := os.WriteFile(testFile, []byte("## Test\n\n```\ncode here\n```\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := linter.Run([]string{testFile})

	if result.TotalErrors == 0 {
		t.Error("expected error for fenced code block without language identifier")
	}
}

func TestRun_EmptyAltText(t *testing.T) {
	cfg := allOff()
	cfg.Rules["empty-alt-text"] = on()

	linter := mustNew(t, cfg)

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
	linter := mustNew(t, cfg)

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

	linter := mustNew(t, cfg)

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

	linter := mustNew(t, cfg)

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

	linter := mustNew(t, cfg)

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

	linter := mustNew(t, cfg)

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

func TestRun_FinalBlankLine_FrontmatterOnly(t *testing.T) {
	cfg := allOff()
	cfg.Rules["final-blank-line"] = on()
	cfg.Rules["no-multiple-blank-lines"] = on()

	linter := mustNew(t, cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "frontmatter_only.md")
	if err := os.WriteFile(testFile, []byte("---\ntitle: \"Docs\"\nweight: 1\n---\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := linter.Run([]string{testFile})

	if result.TotalErrors != 0 {
		t.Errorf("expected no violations for frontmatter-only file, got %d: %v", result.TotalErrors, result.Errors)
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

	linter := mustNew(t, cfg)

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

	linter := mustNew(t, cfg)

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

	linter := mustNew(t, cfg)

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

	linter := mustNew(t, cfg)

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

	linter := mustNew(t, cfg)

	errors, lineCount, _ := linter.LintContent("test.md", "# Title\n\nContent\n")

	if len(errors) == 0 {
		t.Error("expected at least 1 error (heading level)")
	}
	if lineCount != 4 {
		t.Errorf("expected 4 lines, got %d", lineCount)
	}
}

func TestRun_SingleH1(t *testing.T) {
	cfg := allOff()
	cfg.Rules["single-h1"] = on()

	linter := mustNew(t, cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "multi-h1.md")
	if err := os.WriteFile(testFile, []byte("# First\n\n# Second\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := linter.Run([]string{testFile})

	if result.TotalErrors == 0 {
		t.Error("expected error for multiple H1 headings")
	}
}

func TestRun_BlanksAroundHeadings(t *testing.T) {
	cfg := allOff()
	cfg.Rules["blanks-around-headings"] = on()

	linter := mustNew(t, cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "no-blanks.md")
	if err := os.WriteFile(testFile, []byte("Some text\n## Heading\nMore text\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := linter.Run([]string{testFile})

	if result.TotalErrors == 0 {
		t.Error("expected errors for heading without surrounding blank lines")
	}
}

func TestRun_NoBareURLs(t *testing.T) {
	cfg := allOff()
	cfg.Rules["no-bare-urls"] = on()

	linter := mustNew(t, cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "bare-url.md")
	if err := os.WriteFile(testFile, []byte("Visit https://example.com for details.\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := linter.Run([]string{testFile})

	if result.TotalErrors == 0 {
		t.Error("expected error for bare URL")
	}
}

func TestRun_NoEmptyLinks(t *testing.T) {
	cfg := allOff()
	cfg.Rules["no-empty-links"] = on()

	linter := mustNew(t, cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty-link.md")
	if err := os.WriteFile(testFile, []byte("[click here]()\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := linter.Run([]string{testFile})

	if result.TotalErrors == 0 {
		t.Error("expected error for empty link destination")
	}
}

func TestRun_NoEmphasisAsHeading(t *testing.T) {
	cfg := allOff()
	cfg.Rules["no-emphasis-as-heading"] = on()

	linter := mustNew(t, cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "emphasis-heading.md")
	if err := os.WriteFile(testFile, []byte("**Section Title**\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := linter.Run([]string{testFile})

	if result.TotalErrors == 0 {
		t.Error("expected error for emphasis used as heading")
	}
}

func TestRun_BlanksAroundLists(t *testing.T) {
	cfg := allOff()
	cfg.Rules["blanks-around-lists"] = on()

	linter := mustNew(t, cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "lists.md")
	if err := os.WriteFile(testFile, []byte("Some text\n- item 1\n- item 2\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := linter.Run([]string{testFile})

	if result.TotalErrors == 0 {
		t.Error("expected error for list not preceded by blank line")
	}
}

func TestRun_WarningSeverity(t *testing.T) {
	cfg := allOff()
	cfg.Rules["no-setext-headings"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityWarning,
		Options:  map[string]interface{}{},
	}

	linter := mustNew(t, cfg)

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

func TestRun_DisableComment_BlockDisableAll(t *testing.T) {
	cfg := allOff()
	cfg.Rules["no-bare-urls"] = on()

	content := "# Heading\n\n<!-- gomarklint-disable -->\nhttps://example.com\n<!-- gomarklint-enable -->\nhttps://example.com\n"

	lint := mustNew(t, cfg)
	errors, _, _ := lint.LintContent("test.md", content)

	if len(errors) != 1 {
		t.Fatalf("expected 1 error (line 6), got %d: %v", len(errors), errors)
	}
	if errors[0].Line != 6 {
		t.Errorf("expected error on line 6, got line %d", errors[0].Line)
	}
}

func TestRun_DisableComment_BlockDisableNamedRule(t *testing.T) {
	cfg := allOff()
	cfg.Rules["no-bare-urls"] = on()
	cfg.Rules["no-empty-links"] = on()

	content := "<!-- gomarklint-disable no-bare-urls -->\nhttps://example.com\n[]()\n<!-- gomarklint-enable no-bare-urls -->\nhttps://example.com\n"

	lint := mustNew(t, cfg)
	errors, _, _ := lint.LintContent("test.md", content)

	// line 3 (empty link) should still be reported; line 2 (bare URL) should be suppressed
	// line 5 (bare URL after enable) should be reported
	var lines []int
	for _, e := range errors {
		lines = append(lines, e.Line)
	}
	for _, e := range errors {
		if e.Line == 2 && e.Rule == "no-bare-urls" {
			t.Errorf("line 2 no-bare-urls should be suppressed, got %v", e)
		}
	}
	reported := map[int]bool{}
	for _, e := range errors {
		reported[e.Line] = true
	}
	if !reported[3] {
		t.Errorf("line 3 (empty link) should be reported, errors: %v", lines)
	}
	if !reported[5] {
		t.Errorf("line 5 (bare URL after enable) should be reported, errors: %v", lines)
	}
}

func TestRun_DisableComment_DisableLine(t *testing.T) {
	cfg := allOff()
	cfg.Rules["no-bare-urls"] = on()

	content := "https://example.com <!-- gomarklint-disable-line no-bare-urls -->\nhttps://example.com\n"

	lint := mustNew(t, cfg)
	errors, _, _ := lint.LintContent("test.md", content)

	if len(errors) != 1 {
		t.Fatalf("expected 1 error (line 2), got %d: %v", len(errors), errors)
	}
	if errors[0].Line != 2 {
		t.Errorf("expected error on line 2, got line %d", errors[0].Line)
	}
}

func TestRun_DisableComment_DisableNextLine(t *testing.T) {
	cfg := allOff()
	cfg.Rules["no-bare-urls"] = on()

	content := "<!-- gomarklint-disable-next-line no-bare-urls -->\nhttps://example.com\nhttps://example.com\n"

	lint := mustNew(t, cfg)
	errors, _, _ := lint.LintContent("test.md", content)

	if len(errors) != 1 {
		t.Fatalf("expected 1 error (line 3), got %d: %v", len(errors), errors)
	}
	if errors[0].Line != 3 {
		t.Errorf("expected error on line 3, got line %d", errors[0].Line)
	}
}

func TestRun_MaxLineLength_DefaultLimit(t *testing.T) {
	cfg := allOff()
	cfg.Rules["max-line-length"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options:  map[string]interface{}{"lineLength": float64(80)},
	}

	linter := mustNew(t, cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "long.md")
	longLine := "This line is intentionally long and exceeds the eighty character limit set here!!\n"
	if err := os.WriteFile(testFile, []byte(longLine), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := linter.Run([]string{testFile})

	if result.TotalErrors == 0 {
		t.Error("expected error for line exceeding 80 characters")
	}
}

func TestRun_MaxLineLength_CustomLimit(t *testing.T) {
	cfg := allOff()
	cfg.Rules["max-line-length"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options:  map[string]interface{}{"lineLength": float64(120)},
	}

	linter := mustNew(t, cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "long.md")
	// 100 chars — within 120 limit
	shortEnough := make([]byte, 100)
	for i := range shortEnough {
		shortEnough[i] = 'a'
	}
	content := string(shortEnough) + "\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := linter.Run([]string{testFile})

	if result.TotalErrors != 0 {
		t.Errorf("expected no errors for 100-char line with limit=120, got %d", result.TotalErrors)
	}
}

func TestRun_MaxLineLength_NoOptionFallsBackToDefault(t *testing.T) {
	cfg := allOff()
	cfg.Rules["max-line-length"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options:  map[string]interface{}{},
	}

	linter := mustNew(t, cfg)
	if linter.maxLineLength() != 80 {
		t.Errorf("expected default maxLineLength 80, got %d", linter.maxLineLength())
	}
}

func TestRun_DisableComment_NoDisableKeyword_NoOverhead(t *testing.T) {
	cfg := allOff()
	cfg.Rules["no-bare-urls"] = on()

	content := "<!-- some comment -->\nhttps://example.com\n"

	lint := mustNew(t, cfg)
	errors, _, _ := lint.LintContent("test.md", content)

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
}

func TestRun_LinkCheckWithAllowedStatuses(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/allowed":
			w.WriteHeader(http.StatusForbidden) // 403, in allowedStatuses
		case "/blocked":
			w.WriteHeader(http.StatusNotFound) // 404, not in allowedStatuses
		}
	}))
	defer ts.Close()

	cfg := allOff()
	cfg.Rules["external-link"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options: map[string]interface{}{
			"timeoutSeconds":  float64(5),
			"skipPatterns":    []interface{}{},
			"allowedStatuses": []interface{}{float64(403)},
		},
	}

	linter := mustNew(t, cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "links.md")
	content := fmt.Sprintf("[allowed](%s/allowed)\n[blocked](%s/blocked)\n", ts.URL, ts.URL)
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := linter.Run([]string{testFile})

	if result.TotalErrors != 1 {
		t.Errorf("expected 1 error (404 only), got %d", result.TotalErrors)
	}
}

func TestRun_NoTrailingPunctuation_Violation(t *testing.T) {
	cfg := allOff()
	cfg.Rules["no-trailing-punctuation"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options:  map[string]interface{}{"punctuation": config.DefaultNoTrailingPunctuation},
	}

	lint := mustNew(t, cfg)
	errors, _, _ := lint.LintContent("test.md", "## Heading.\n")

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
}

func TestRun_NoTrailingPunctuation_NoOptionFallsBackToDefault(t *testing.T) {
	cfg := allOff()
	cfg.Rules["no-trailing-punctuation"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options:  map[string]interface{}{},
	}

	linter := mustNew(t, cfg)
	if linter.noTrailingPunctuation() != config.DefaultNoTrailingPunctuation {
		t.Errorf("expected default punctuation %q, got %q", config.DefaultNoTrailingPunctuation, linter.noTrailingPunctuation())
	}
}

func TestRun_LinkFragments_DetectsViolation(t *testing.T) {
	cfg := allOff()
	cfg.Rules["link-fragments"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options:  map[string]interface{}{"slug-algorithm": "github"},
	}

	lint := mustNew(t, cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	content := "## Introduction\n\nSee [Setup](#setup) for details.\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := lint.Run([]string{testFile})

	if result.TotalErrors == 0 {
		t.Error("expected error for broken fragment link")
	}
}

func TestRun_LinkFragments_ValidLink(t *testing.T) {
	cfg := allOff()
	cfg.Rules["link-fragments"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options:  map[string]interface{}{"slug-algorithm": "github"},
	}

	lint := mustNew(t, cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	content := "## Introduction\n\nSee [Introduction](#introduction) for details.\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := lint.Run([]string{testFile})

	if result.TotalErrors != 0 {
		t.Errorf("expected no errors for valid fragment link, got %d", result.TotalErrors)
	}
}

func TestRun_ConsistentCodeFence_Violation(t *testing.T) {
	cfg := allOff()
	cfg.Rules["consistent-code-fence"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options:  map[string]interface{}{"style": "consistent"},
	}

	lint := mustNew(t, cfg)
	errors, _, _ := lint.LintContent("test.md", "```go\ncode\n```\n\n~~~python\ncode\n~~~\n")

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
}

func TestRun_ConsistentCodeFence_NoOptionFallsBackToConsistent(t *testing.T) {
	cfg := allOff()
	cfg.Rules["consistent-code-fence"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options:  map[string]interface{}{},
	}

	linter := mustNew(t, cfg)
	if linter.consistentCodeFenceStyle() != "consistent" {
		t.Errorf("expected fallback style %q, got %q", "consistent", linter.consistentCodeFenceStyle())
	}
}

func TestRun_ConsistentEmphasisStyle_Violation(t *testing.T) {
	cfg := allOff()
	cfg.Rules["consistent-emphasis-style"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options:  map[string]interface{}{"style": "asterisk"},
	}

	lint := mustNew(t, cfg)
	errors, _, _ := lint.LintContent("test.md", "This is _italic_ text.\n")

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
}

func TestRun_ConsistentListMarker_Violation(t *testing.T) {
	cfg := allOff()
	cfg.Rules["consistent-list-marker"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options:  map[string]interface{}{"style": "dash"},
	}

	lint := mustNew(t, cfg)
	errors, _, _ := lint.LintContent("test.md", "* item\n")

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
}

func TestRun_ConsistentEmphasisStyle_NoOptionFallsBackToConsistent(t *testing.T) {
	cfg := allOff()
	cfg.Rules["consistent-emphasis-style"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options:  map[string]interface{}{},
	}

	linter := mustNew(t, cfg)
	if linter.consistentEmphasisStyle() != "consistent" {
		t.Errorf("expected fallback style %q, got %q", "consistent", linter.consistentEmphasisStyle())
	}
}

func TestRun_ConsistentListMarker_NoOptionFallsBackToConsistent(t *testing.T) {
	cfg := allOff()
	cfg.Rules["consistent-list-marker"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options:  map[string]interface{}{},
	}

	linter := mustNew(t, cfg)
	if linter.consistentListMarkerStyle() != "consistent" {
		t.Errorf("expected fallback style %q, got %q", "consistent", linter.consistentListMarkerStyle())
	}
}

func TestNew_InvalidStyleOption_ConsistentCodeFence(t *testing.T) {
	cfg := allOff()
	cfg.Rules["consistent-code-fence"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options:  map[string]interface{}{"style": "hoge"},
	}

	_, err := New(cfg)
	if err == nil {
		t.Fatal("expected error for invalid style, got nil")
	}
	want := `gomarklint: invalid value "hoge" for consistent-code-fence.style (valid values: consistent, backtick, tilde)`
	if err.Error() != want {
		t.Errorf("unexpected error message:\ngot:  %s\nwant: %s", err.Error(), want)
	}
}

func TestNew_InvalidStyleOption_ConsistentEmphasisStyle(t *testing.T) {
	cfg := allOff()
	cfg.Rules["consistent-emphasis-style"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options:  map[string]interface{}{"style": "bold"},
	}

	_, err := New(cfg)
	if err == nil {
		t.Fatal("expected error for invalid style, got nil")
	}
	want := `gomarklint: invalid value "bold" for consistent-emphasis-style.style (valid values: consistent, asterisk, underscore)`
	if err.Error() != want {
		t.Errorf("unexpected error message:\ngot:  %s\nwant: %s", err.Error(), want)
	}
}

func TestNew_InvalidStyleOption_ConsistentListMarker(t *testing.T) {
	cfg := allOff()
	cfg.Rules["consistent-list-marker"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options:  map[string]interface{}{"style": "bullet"},
	}

	_, err := New(cfg)
	if err == nil {
		t.Fatal("expected error for invalid style, got nil")
	}
	want := `gomarklint: invalid value "bullet" for consistent-list-marker.style (valid values: consistent, dash, asterisk, plus)`
	if err.Error() != want {
		t.Errorf("unexpected error message:\ngot:  %s\nwant: %s", err.Error(), want)
	}
}

func TestNew_InvalidStyleOption_NonStringValue(t *testing.T) {
	cfg := allOff()
	cfg.Rules["consistent-code-fence"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options:  map[string]interface{}{"style": 123},
	}

	_, err := New(cfg)
	if err == nil {
		t.Fatal("expected error for non-string style value, got nil")
	}
}

func TestNew_InvalidStyleOption_DisabledRuleStillValidated(t *testing.T) {
	cfg := allOff()
	cfg.Rules["consistent-code-fence"] = &config.RuleConfig{
		Enabled:  false,
		Severity: config.SeverityError,
		Options:  map[string]interface{}{"style": "hoge"},
	}

	_, err := New(cfg)
	if err == nil {
		t.Fatal("expected error for invalid style even when rule is disabled, got nil")
	}
}

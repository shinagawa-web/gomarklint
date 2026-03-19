package app

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/shinagawa-web/gomarklint/v2/internal/config"
	"github.com/shinagawa-web/gomarklint/v2/internal/linter"
	"github.com/shinagawa-web/gomarklint/v2/internal/rule"
)

func TestRun_ValidMarkdown(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "valid.md")
	os.WriteFile(f, []byte("## Hello\n\nWorld.\n"), 0644)

	var buf bytes.Buffer
	err := Run(&buf, Options{
		ConfigPath: "/nonexistent/.gomarklint.json",
		Args:       []string{f},
	})
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestRun_InvalidMarkdown_ReturnsViolations(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "invalid.md")
	os.WriteFile(f, []byte("# H1 heading\n"), 0644) // H1 when default minLevel=2

	var buf bytes.Buffer
	err := Run(&buf, Options{
		ConfigPath: "/nonexistent/.gomarklint.json",
		Args:       []string{f},
	})
	if !errors.Is(err, ErrLintViolations) {
		t.Errorf("expected ErrLintViolations, got: %v", err)
	}
}

func TestRun_NoArgs_NoInclude_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "empty-include.json")
	os.WriteFile(cfgFile, []byte(`{"default":true,"rules":{},"include":[]}`), 0644)

	var buf bytes.Buffer
	err := Run(&buf, Options{
		ConfigPath: cfgFile,
		Args:       []string{},
	})
	if err == nil || !strings.Contains(err.Error(), "please provide") {
		t.Errorf("expected 'please provide' error, got: %v", err)
	}
}

func TestRun_InvalidConfigFile_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	badConfig := filepath.Join(dir, "bad.json")
	os.WriteFile(badConfig, []byte("{invalid json}"), 0644)

	var buf bytes.Buffer
	err := Run(&buf, Options{
		ConfigPath: badConfig,
		Args:       []string{"somefile.md"},
	})
	if err == nil {
		t.Error("expected error for invalid config, got nil")
	}
}

func TestRun_InvalidOutputFormat_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "bad-output.json")
	os.WriteFile(cfgFile, []byte(`{"default":true,"rules":{},"output":"xlsx"}`), 0644)

	var buf bytes.Buffer
	err := Run(&buf, Options{
		ConfigPath: cfgFile,
		Args:       []string{"somefile.md"},
	})
	if err == nil || !strings.Contains(err.Error(), "invalid output format") {
		t.Errorf("expected 'invalid output format' error, got: %v", err)
	}
}

func TestRun_OutputFormatOverride(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "valid.md")
	os.WriteFile(f, []byte("## Hello\n\nWorld.\n"), 0644)

	var buf bytes.Buffer
	err := Run(&buf, Options{
		ConfigPath:   "/nonexistent/.gomarklint.json",
		Args:         []string{f},
		OutputFormat: "json",
	})
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if !strings.Contains(buf.String(), `"files"`) {
		t.Errorf("expected JSON output, got: %s", buf.String())
	}
}

func TestRun_MinSeverityOverride(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "warning-rule.json")
	os.WriteFile(cfgFile, []byte(`{
		"default": false,
		"rules": {
			"no-setext-headings": "warning"
		}
	}`), 0644)
	f := filepath.Join(dir, "setext.md")
	os.WriteFile(f, []byte("Heading\n=======\n\nParagraph.\n"), 0644)

	// With MinSeverity=warning, warning violations are shown
	var buf bytes.Buffer
	err := Run(&buf, Options{
		ConfigPath:  cfgFile,
		Args:        []string{f},
		MinSeverity: "warning",
	})
	if err != nil {
		t.Errorf("expected no error (warnings don't fail), got: %v", err)
	}
	if !strings.Contains(buf.String(), "warning") {
		t.Errorf("expected warning in output, got: %s", buf.String())
	}

	// With MinSeverity=error, warnings are suppressed
	buf.Reset()
	err = Run(&buf, Options{
		ConfigPath:  cfgFile,
		Args:        []string{f},
		MinSeverity: "error",
	})
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if strings.Contains(buf.String(), "warning") {
		t.Errorf("expected warning suppressed, got: %s", buf.String())
	}
}

func TestRun_UsesConfigInclude(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "valid.md")
	os.WriteFile(f, []byte("## Hello\n\nWorld.\n"), 0644)

	cfgFile := filepath.Join(dir, "include.json")
	os.WriteFile(cfgFile, []byte(`{"default":true,"rules":{},"include":["`+f+`"]}`), 0644)

	var buf bytes.Buffer
	// No args — should fall back to cfg.Include
	err := Run(&buf, Options{
		ConfigPath: cfgFile,
		Args:       []string{},
	})
	if err != nil {
		t.Errorf("expected no error when using config include, got: %v", err)
	}
}

type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (int, error) {
	return 0, errors.New("write error")
}

func TestFormatOutput_WriterError(t *testing.T) {
	cfg := config.Default()
	cfg.Rules["external-link"].Enabled = false
	result := &linter.Result{
		Errors:       map[string][]rule.LintError{},
		OrderedPaths: []string{},
	}
	err := formatOutput(&errorWriter{}, cfg, result, 1, time.Millisecond)
	if err == nil {
		t.Error("expected error from bad writer, got nil")
	}
}

func TestRun_WriterError(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "valid.md")
	os.WriteFile(f, []byte("## Hello\n\nWorld.\n"), 0644)

	err := Run(&errorWriter{}, Options{
		ConfigPath: "/nonexistent/.gomarklint.json",
		Args:       []string{f},
	})
	if err == nil {
		t.Error("expected error from bad writer, got nil")
	}
}

func TestFilterBySeverity(t *testing.T) {
	errViolation := rule.LintError{File: "a.md", Line: 1, Message: "error violation", Severity: "error"}
	warnViolation := rule.LintError{File: "a.md", Line: 2, Message: "warn violation", Severity: "warning"}

	t.Run("MinSeverityWarning_ShowsAll", func(t *testing.T) {
		details := map[string][]rule.LintError{
			"a.md": {errViolation, warnViolation},
		}
		filtered, errCount, warnCount := filterBySeverity(details, config.SeverityWarning)
		if errCount != 1 || warnCount != 1 {
			t.Errorf("expected 1 error and 1 warning, got %d errors %d warnings", errCount, warnCount)
		}
		if len(filtered["a.md"]) != 2 {
			t.Errorf("expected 2 violations in filtered, got %d", len(filtered["a.md"]))
		}
	})

	t.Run("MinSeverityError_SuppressesWarnings", func(t *testing.T) {
		details := map[string][]rule.LintError{
			"a.md": {errViolation, warnViolation},
		}
		filtered, errCount, warnCount := filterBySeverity(details, config.SeverityError)
		if errCount != 1 || warnCount != 0 {
			t.Errorf("expected 1 error and 0 warnings, got %d errors %d warnings", errCount, warnCount)
		}
		if len(filtered["a.md"]) != 1 || filtered["a.md"][0].Severity != "error" {
			t.Errorf("expected only error violation in filtered, got %v", filtered["a.md"])
		}
	})

	t.Run("MinSeverityError_WarningsOnlyFile_OmitsPath", func(t *testing.T) {
		details := map[string][]rule.LintError{
			"a.md": {warnViolation},
		}
		filtered, errCount, warnCount := filterBySeverity(details, config.SeverityError)
		if errCount != 0 || warnCount != 0 {
			t.Errorf("expected 0 errors and 0 warnings, got %d errors %d warnings", errCount, warnCount)
		}
		if _, exists := filtered["a.md"]; exists {
			t.Error("expected path to be omitted from filtered map when no violations remain")
		}
	})

	t.Run("MultipleFiles_PartialFilter", func(t *testing.T) {
		details := map[string][]rule.LintError{
			"a.md": {warnViolation},
			"b.md": {errViolation},
		}
		filtered, errCount, warnCount := filterBySeverity(details, config.SeverityError)
		if errCount != 1 || warnCount != 0 {
			t.Errorf("expected 1 error and 0 warnings, got %d errors %d warnings", errCount, warnCount)
		}
		if _, exists := filtered["a.md"]; exists {
			t.Error("warning-only path should be omitted")
		}
		if len(filtered["b.md"]) != 1 {
			t.Errorf("expected 1 violation for b.md, got %d", len(filtered["b.md"]))
		}
	})

	t.Run("EmptyDetails", func(t *testing.T) {
		filtered, errCount, warnCount := filterBySeverity(map[string][]rule.LintError{}, config.SeverityWarning)
		if errCount != 0 || warnCount != 0 || len(filtered) != 0 {
			t.Errorf("expected empty result, got errCount=%d warnCount=%d filtered=%v", errCount, warnCount, filtered)
		}
	})
}

func TestFormatOutput(t *testing.T) {
	cfg := config.Default()
	cfg.Rules["external-link"].Enabled = false

	t.Run("TextFormat_NoViolations", func(t *testing.T) {
		result := &linter.Result{
			Errors:       map[string][]rule.LintError{},
			OrderedPaths: []string{},
		}
		var buf bytes.Buffer
		err := formatOutput(&buf, cfg, result, 1, time.Millisecond)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(buf.String(), "No issues found") {
			t.Errorf("expected 'No issues found', got: %s", buf.String())
		}
	})

	t.Run("TextFormat_WithErrors", func(t *testing.T) {
		result := &linter.Result{
			Errors: map[string][]rule.LintError{
				"a.md": {{File: "a.md", Line: 1, Message: "test error", Severity: "error"}},
			},
			OrderedPaths: []string{"a.md"},
			TotalErrors:  1,
		}
		var buf bytes.Buffer
		err := formatOutput(&buf, cfg, result, 1, time.Millisecond)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(buf.String(), "issues found") {
			t.Errorf("expected 'issues found', got: %s", buf.String())
		}
	})

	t.Run("TextFormat_WithWarnings", func(t *testing.T) {
		result := &linter.Result{
			Errors: map[string][]rule.LintError{
				"a.md": {{File: "a.md", Line: 1, Message: "setext heading", Severity: "warning"}},
			},
			OrderedPaths:  []string{"a.md"},
			TotalWarnings: 1,
		}
		var buf bytes.Buffer
		err := formatOutput(&buf, cfg, result, 1, time.Millisecond)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(buf.String(), "warning") {
			t.Errorf("expected 'warning' in output, got: %s", buf.String())
		}
	})

	t.Run("JSONFormat", func(t *testing.T) {
		cfgJSON := config.Default()
		cfgJSON.Rules["external-link"].Enabled = false
		cfgJSON.OutputFormat = "json"

		result := &linter.Result{
			Errors:       map[string][]rule.LintError{},
			OrderedPaths: []string{},
		}
		var buf bytes.Buffer
		err := formatOutput(&buf, cfgJSON, result, 1, time.Millisecond)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(buf.String(), `"files"`) {
			t.Errorf("expected JSON output, got: %s", buf.String())
		}
	})

	t.Run("ExternalLinkEnabled_ShowsLinkCount", func(t *testing.T) {
		cfgLink := config.Default()
		cfgLink.Rules["external-link"].Enabled = true
		result := &linter.Result{
			Errors:            map[string][]rule.LintError{},
			OrderedPaths:      []string{},
			TotalLinksChecked: 5,
		}
		var buf bytes.Buffer
		err := formatOutput(&buf, cfgLink, result, 1, time.Millisecond)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(buf.String(), "link(s)") {
			t.Errorf("expected link count in output, got: %s", buf.String())
		}
	})
}

package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/shinagawa-web/gomarklint/v2/internal/rule"
)

func TestTextFormatter_NoErrors(t *testing.T) {
	formatter := NewTextFormatter()
	result := &Result{
		Files:        3,
		Lines:        150,
		Total:        0,
		LinksChecked: nil,
		Duration:     500 * time.Millisecond,
		Details:      map[string][]rule.LintError{},
		OrderedPaths: []string{},
	}

	var buf bytes.Buffer
	err := formatter.Format(&buf, result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No issues found") {
		t.Errorf("expected 'No issues found' in output, got: %s", output)
	}
	if !strings.Contains(output, "Checked 3 file(s), 150 line(s)") {
		t.Errorf("expected file and line count in output, got: %s", output)
	}
	if !strings.Contains(output, "500ms") {
		t.Errorf("expected duration in ms, got: %s", output)
	}
}

func TestTextFormatter_WithErrors(t *testing.T) {
	formatter := NewTextFormatter()
	result := &Result{
		Files: 2,
		Lines: 100,
		Total: 2,
		Details: map[string][]rule.LintError{
			"file1.md": {
				{File: "file1.md", Line: 5, Message: "Heading level error"},
			},
			"file2.md": {
				{File: "file2.md", Line: 10, Message: "Missing blank line"},
			},
		},
		OrderedPaths: []string{"file1.md", "file2.md"},
		Duration:     1500 * time.Millisecond,
	}

	output := formatAndGetOutput(t, formatter, result)

	assertContains(t, output, "2 issues found")
	assertContains(t, output, "Errors in file1.md:")
	assertContains(t, output, "file1.md:5: [error] Heading level error")
	assertContains(t, output, "Errors in file2.md:")
	assertContains(t, output, "file2.md:10: [error] Missing blank line")
	assertContains(t, output, "1.5s")
}

func TestTextFormatter_WithLinkCheck(t *testing.T) {
	formatter := NewTextFormatter()
	linksChecked := 25
	result := &Result{
		Files:        5,
		Lines:        200,
		Total:        0,
		LinksChecked: &linksChecked,
		Duration:     2 * time.Second,
		Details:      map[string][]rule.LintError{},
		OrderedPaths: []string{},
	}

	output := formatAndGetOutput(t, formatter, result)

	assertContains(t, output, "25 link(s)")
	assertContains(t, output, "Checked 5 file(s), 200 line(s), 25 link(s)")
}

func TestTextFormatter_WithMixedErrorsAndEmptyFiles(t *testing.T) {
	formatter := NewTextFormatter()
	result := &Result{
		Files: 3,
		Lines: 150,
		Total: 1,
		Details: map[string][]rule.LintError{
			"file1.md": {
				{File: "file1.md", Line: 5, Message: "Heading error"},
			},
			"file2.md": {},
			"file3.md": {},
		},
		OrderedPaths: []string{"file1.md", "file2.md", "file3.md"},
		Duration:     800 * time.Millisecond,
	}

	output := formatAndGetOutput(t, formatter, result)

	assertContains(t, output, "Errors in file1.md:")
	assertNotContains(t, output, "Errors in file2.md:")
	assertNotContains(t, output, "Errors in file3.md:")
	assertContains(t, output, "1 issues found")
}

func formatAndGetOutput(t *testing.T, formatter Formatter, result *Result) string {
	t.Helper()
	var buf bytes.Buffer
	err := formatter.Format(&buf, result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return buf.String()
}

func assertContains(t *testing.T, output, expected string) {
	t.Helper()
	if !strings.Contains(output, expected) {
		t.Errorf("expected '%s' in output, got: %s", expected, output)
	}
}

func assertNotContains(t *testing.T, output, unexpected string) {
	t.Helper()
	if strings.Contains(output, unexpected) {
		t.Errorf("should not contain '%s' in output, got: %s", unexpected, output)
	}
}

func TestTextFormatter_WarningsOnly(t *testing.T) {
	formatter := NewTextFormatter()

	t.Run("SingularWarning", func(t *testing.T) {
		result := &Result{
			Files:    1,
			Lines:    10,
			Total:    1,
			Warnings: 1,
			Details: map[string][]rule.LintError{
				"file.md": {
					{File: "file.md", Line: 5, Message: "Setext heading found", Severity: "warning"},
				},
			},
			OrderedPaths: []string{"file.md"},
			Duration:     100 * time.Millisecond,
		}

		output := formatAndGetOutput(t, formatter, result)

		assertContains(t, output, "[warning]")
		assertContains(t, output, "1 warning found")
		assertNotContains(t, output, "issues found")
		assertNotContains(t, output, "[error]")
	})

	t.Run("PluralWarnings", func(t *testing.T) {
		result := &Result{
			Files:    1,
			Lines:    10,
			Total:    2,
			Warnings: 2,
			Details: map[string][]rule.LintError{
				"file.md": {
					{File: "file.md", Line: 5, Message: "Setext heading", Severity: "warning"},
					{File: "file.md", Line: 8, Message: "Another warning", Severity: "warning"},
				},
			},
			OrderedPaths: []string{"file.md"},
			Duration:     100 * time.Millisecond,
		}

		output := formatAndGetOutput(t, formatter, result)

		assertContains(t, output, "2 warnings found")
		assertNotContains(t, output, "issues found")
	})

	t.Run("WriteErrorOnWarningSummary", func(t *testing.T) {
		result := &Result{
			Files:        1,
			Lines:        10,
			Total:        1,
			Warnings:     1,
			Details:      map[string][]rule.LintError{},
			OrderedPaths: []string{},
			Duration:     100 * time.Millisecond,
		}

		// limit=1 lets the leading "\n" through, then fails on the warning summary write
		lw := &limitedErrorWriter{limit: 1}
		err := formatter.Format(lw, result)
		if err == nil {
			t.Error("expected error when writing warning summary")
		}
	})
}

func TestTextFormatter_WriteErrors(t *testing.T) {
	t.Run("ErrorInErrorDetailsGeneral", func(t *testing.T) {
		formatter := NewTextFormatter()
		result := &Result{
			Files: 1,
			Lines: 10,
			Total: 1,
			Details: map[string][]rule.LintError{
				"test.md": {
					{File: "test.md", Line: 5, Message: "Test error"},
				},
			},
			OrderedPaths: []string{"test.md"},
			Duration:     100 * time.Millisecond,
		}

		ew := &errorWriter{}
		err := formatter.Format(ew, result)
		if err == nil {
			t.Error("expected error when writing to errorWriter")
		}
	})

	t.Run("ErrorInSummary", func(t *testing.T) {
		formatter := NewTextFormatter()
		result := &Result{
			Files:        1,
			Lines:        10,
			Total:        0,
			Duration:     100 * time.Millisecond,
			Details:      map[string][]rule.LintError{},
			OrderedPaths: []string{},
		}

		var buf bytes.Buffer

		// First write should succeed
		err := formatter.Format(&buf, result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// limit=1 lets the leading "\n" through, then fails on the "No issues found" write
		lw := &limitedErrorWriter{limit: 1}
		err = formatter.Format(lw, result)
		if err == nil {
			t.Error("expected error when writing summary")
		}
	})

	t.Run("ErrorInStats", func(t *testing.T) {
		formatter := NewTextFormatter()
		linksChecked := 10
		result := &Result{
			Files:        1,
			Lines:        10,
			Total:        0,
			LinksChecked: &linksChecked,
			Duration:     100 * time.Millisecond,
			Details:      map[string][]rule.LintError{},
			OrderedPaths: []string{},
		}

		// Allow formatSummary to succeed (about 31 bytes incl. leading \n) but fail on formatStats
		lw := &limitedErrorWriter{limit: 31}
		err := formatter.Format(lw, result)
		if err == nil {
			t.Error("expected error when writing stats to errorWriter")
		}
	})

	t.Run("ErrorInErrorDetailsHeader", func(t *testing.T) {
		formatter := NewTextFormatter()
		result := &Result{
			Files: 1,
			Lines: 10,
			Total: 1,
			Details: map[string][]rule.LintError{
				"test.md": {
					{File: "test.md", Line: 5, Message: "Test error"},
				},
			},
			OrderedPaths: []string{"test.md"},
			Duration:     100 * time.Millisecond,
		}

		// limit=1 lets the leading "\n" through, then fails on "Errors in test.md:\n"
		lw := &limitedErrorWriter{limit: 1}
		err := formatter.Format(lw, result)
		if err == nil {
			t.Error("expected error when writing error details header")
		}
	})

	t.Run("ErrorInErrorDetailsLine", func(t *testing.T) {
		formatter := NewTextFormatter()
		result := &Result{
			Files: 1,
			Lines: 10,
			Total: 1,
			Details: map[string][]rule.LintError{
				"test.md": {
					{File: "test.md", Line: 5, Message: "Test error"},
				},
			},
			OrderedPaths: []string{"test.md"},
			Duration:     100 * time.Millisecond,
		}

		// Allow "\n" (1) + "Errors in test.md:\n" (19) = 20 bytes, fail on error line
		lw := &limitedErrorWriter{limit: 20}
		err := formatter.Format(lw, result)
		if err == nil {
			t.Error("expected error when writing error line")
		}
	})

	t.Run("ErrorInErrorDetailsNewline", func(t *testing.T) {
		formatter := NewTextFormatter()
		result := &Result{
			Files: 1,
			Lines: 10,
			Total: 1,
			Details: map[string][]rule.LintError{
				"test.md": {
					{File: "test.md", Line: 5, Message: "Test error"},
				},
			},
			OrderedPaths: []string{"test.md"},
			Duration:     100 * time.Millisecond,
		}

		// Allow "\n" (1) + header (19 bytes) + detail line (32 bytes) = 52, fail on final newline
		// Detail line: "  test.md:5: [error] Test error\n" = 32 bytes
		lw := &limitedErrorWriter{limit: 52}
		err := formatter.Format(lw, result)
		if err == nil {
			t.Error("expected error when writing final newline after errors")
		}
	})

	t.Run("ErrorInSummaryWithErrors", func(t *testing.T) {
		formatter := NewTextFormatter()
		result := &Result{
			Files:        1,
			Lines:        10,
			Total:        1,
			Details:      map[string][]rule.LintError{},
			OrderedPaths: []string{},
			Duration:     100 * time.Millisecond,
		}

		// limit=1 lets the leading "\n" through, then fails on the "N issues found" write
		lw := &limitedErrorWriter{limit: 1}
		err := formatter.Format(lw, result)
		if err == nil {
			t.Error("expected error when writing summary with errors")
		}
	})
}

func TestTextFormatter_StatsFormatting(t *testing.T) {
	t.Run("WithLinksShortDuration", func(t *testing.T) {
		formatter := NewTextFormatter()
		linksChecked := 5
		result := &Result{
			Files:        2,
			Lines:        50,
			Total:        0,
			LinksChecked: &linksChecked,
			Duration:     500 * time.Millisecond,
			Details:      map[string][]rule.LintError{},
			OrderedPaths: []string{},
		}

		var buf bytes.Buffer
		err := formatter.Format(&buf, result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "500ms") {
			t.Errorf("expected 500ms duration, got: %s", output)
		}
		if !strings.Contains(output, "5 link(s)") {
			t.Errorf("expected 5 links checked, got: %s", output)
		}
	})

	t.Run("WithLinksLongDuration", func(t *testing.T) {
		formatter := NewTextFormatter()
		linksChecked := 10
		result := &Result{
			Files:        3,
			Lines:        100,
			Total:        0,
			LinksChecked: &linksChecked,
			Duration:     2500 * time.Millisecond,
			Details:      map[string][]rule.LintError{},
			OrderedPaths: []string{},
		}

		var buf bytes.Buffer
		err := formatter.Format(&buf, result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "2.5s") {
			t.Errorf("expected 2.5s duration, got: %s", output)
		}
		if !strings.Contains(output, "10 link(s)") {
			t.Errorf("expected 10 links checked, got: %s", output)
		}
	})

	t.Run("WithoutLinksShortDuration", func(t *testing.T) {
		formatter := NewTextFormatter()
		result := &Result{
			Files:        1,
			Lines:        25,
			Total:        0,
			LinksChecked: nil,
			Duration:     300 * time.Millisecond,
			Details:      map[string][]rule.LintError{},
			OrderedPaths: []string{},
		}

		var buf bytes.Buffer
		err := formatter.Format(&buf, result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "300ms") {
			t.Errorf("expected 300ms duration, got: %s", output)
		}
		if strings.Contains(output, "link(s)") {
			t.Errorf("should not mention links when not checked, got: %s", output)
		}
	})

	t.Run("WithoutLinksLongDuration", func(t *testing.T) {
		formatter := NewTextFormatter()
		result := &Result{
			Files:        2,
			Lines:        75,
			Total:        0,
			LinksChecked: nil,
			Duration:     1500 * time.Millisecond,
			Details:      map[string][]rule.LintError{},
			OrderedPaths: []string{},
		}

		var buf bytes.Buffer
		err := formatter.Format(&buf, result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "1.5s") {
			t.Errorf("expected 1.5s duration, got: %s", output)
		}
		if strings.Contains(output, "link(s)") {
			t.Errorf("should not mention links when not checked, got: %s", output)
		}
	})
}

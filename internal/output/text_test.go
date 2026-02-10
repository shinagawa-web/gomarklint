package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/shinagawa-web/gomarklint/internal/rule"
)

func TestTextFormatter_Format(t *testing.T) {
	t.Run("NoErrors", func(t *testing.T) {
		formatter := NewTextFormatter()
		result := &Result{
			Files:        3,
			Lines:        150,
			Errors:       0,
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
	})

	t.Run("WithErrors", func(t *testing.T) {
		formatter := NewTextFormatter()
		result := &Result{
			Files:  2,
			Lines:  100,
			Errors: 2,
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

		var buf bytes.Buffer
		err := formatter.Format(&buf, result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "2 issues found") {
			t.Errorf("expected '2 issues found' in output, got: %s", output)
		}
		if !strings.Contains(output, "Errors in file1.md:") {
			t.Errorf("expected file1.md errors section, got: %s", output)
		}
		if !strings.Contains(output, "file1.md:5: Heading level error") {
			t.Errorf("expected file1.md error detail, got: %s", output)
		}
		if !strings.Contains(output, "Errors in file2.md:") {
			t.Errorf("expected file2.md errors section, got: %s", output)
		}
		if !strings.Contains(output, "file2.md:10: Missing blank line") {
			t.Errorf("expected file2.md error detail, got: %s", output)
		}
		if !strings.Contains(output, "1.5s") {
			t.Errorf("expected duration in seconds, got: %s", output)
		}
	})

	t.Run("WithLinkCheck", func(t *testing.T) {
		formatter := NewTextFormatter()
		linksChecked := 25
		result := &Result{
			Files:        5,
			Lines:        200,
			Errors:       0,
			LinksChecked: &linksChecked,
			Duration:     2 * time.Second,
			Details:      map[string][]rule.LintError{},
			OrderedPaths: []string{},
		}

		var buf bytes.Buffer
		err := formatter.Format(&buf, result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "25 link(s)") {
			t.Errorf("expected link count in output, got: %s", output)
		}
		if !strings.Contains(output, "Checked 5 file(s), 200 line(s), 25 link(s)") {
			t.Errorf("expected full stats with links, got: %s", output)
		}
	})

	t.Run("WithMixedErrorsAndEmptyFiles", func(t *testing.T) {
		formatter := NewTextFormatter()
		result := &Result{
			Files:  3,
			Lines:  150,
			Errors: 1,
			Details: map[string][]rule.LintError{
				"file1.md": {
					{File: "file1.md", Line: 5, Message: "Heading error"},
				},
				"file2.md": {}, // empty errors - should be skipped (continue)
				"file3.md": {}, // empty errors - should be skipped (continue)
			},
			OrderedPaths: []string{"file1.md", "file2.md", "file3.md"},
			Duration:     800 * time.Millisecond,
		}

		var buf bytes.Buffer
		err := formatter.Format(&buf, result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Errors in file1.md:") {
			t.Errorf("expected file1.md errors section, got: %s", output)
		}
		if strings.Contains(output, "Errors in file2.md:") {
			t.Errorf("should not show file2.md (no errors), got: %s", output)
		}
		if strings.Contains(output, "Errors in file3.md:") {
			t.Errorf("should not show file3.md (no errors), got: %s", output)
		}
		if !strings.Contains(output, "1 issues found") {
			t.Errorf("expected '1 issues found' in output, got: %s", output)
		}
	})
}

func TestTextFormatter_WriteErrors(t *testing.T) {
	t.Run("ErrorInErrorDetailsGeneral", func(t *testing.T) {
		formatter := NewTextFormatter()
		result := &Result{
			Files:  1,
			Lines:  10,
			Errors: 1,
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
			Errors:       0,
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

		// Writing to error writer should fail
		ew := &errorWriter{}
		err = formatter.Format(ew, result)
		if err == nil {
			t.Error("expected error when writing summary to errorWriter")
		}
	})

	t.Run("ErrorInStats", func(t *testing.T) {
		formatter := NewTextFormatter()
		linksChecked := 10
		result := &Result{
			Files:        1,
			Lines:        10,
			Errors:       0,
			LinksChecked: &linksChecked,
			Duration:     100 * time.Millisecond,
			Details:      map[string][]rule.LintError{},
			OrderedPaths: []string{},
		}

		// Allow formatSummary to succeed (about 30 bytes) but fail on formatStats
		lw := &limitedErrorWriter{limit: 30}
		err := formatter.Format(lw, result)
		if err == nil {
			t.Error("expected error when writing stats to errorWriter")
		}
	})

	t.Run("ErrorInErrorDetailsHeader", func(t *testing.T) {
		formatter := NewTextFormatter()
		result := &Result{
			Files:  1,
			Lines:  10,
			Errors: 1,
			Details: map[string][]rule.LintError{
				"test.md": {
					{File: "test.md", Line: 5, Message: "Test error"},
				},
			},
			OrderedPaths: []string{"test.md"},
			Duration:     100 * time.Millisecond,
		}

		// Fail immediately when trying to write "Errors in test.md:"
		ew := &errorWriter{}
		err := formatter.Format(ew, result)
		if err == nil {
			t.Error("expected error when writing error details header")
		}
	})

	t.Run("ErrorInErrorDetailsLine", func(t *testing.T) {
		formatter := NewTextFormatter()
		result := &Result{
			Files:  1,
			Lines:  10,
			Errors: 1,
			Details: map[string][]rule.LintError{
				"test.md": {
					{File: "test.md", Line: 5, Message: "Test error"},
				},
			},
			OrderedPaths: []string{"test.md"},
			Duration:     100 * time.Millisecond,
		}

		// Allow "Errors in test.md:\n" (19 bytes) but fail on error line
		lw := &limitedErrorWriter{limit: 19}
		err := formatter.Format(lw, result)
		if err == nil {
			t.Error("expected error when writing error line")
		}
	})

	t.Run("ErrorInErrorDetailsNewline", func(t *testing.T) {
		formatter := NewTextFormatter()
		result := &Result{
			Files:  1,
			Lines:  10,
			Errors: 1,
			Details: map[string][]rule.LintError{
				"test.md": {
					{File: "test.md", Line: 5, Message: "Test error"},
				},
			},
			OrderedPaths: []string{"test.md"},
			Duration:     100 * time.Millisecond,
		}

		// Allow header and error line (43 bytes) but fail on final newline
		lw := &limitedErrorWriter{limit: 43}
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
			Errors:       1,
			Details:      map[string][]rule.LintError{},
			OrderedPaths: []string{},
			Duration:     100 * time.Millisecond,
		}

		ew := &errorWriter{}
		err := formatter.Format(ew, result)
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
			Errors:       0,
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
			Errors:       0,
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
			Errors:       0,
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
			Errors:       0,
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

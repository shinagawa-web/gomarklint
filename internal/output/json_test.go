package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/shinagawa-web/gomarklint/internal/rule"
)

func TestJSONFormatter_Format(t *testing.T) {
	t.Run("NoErrors", func(t *testing.T) {
		formatter := NewJSONFormatter()
		result := &Result{
			Files:        2,
			Lines:        100,
			Errors:       0,
			LinksChecked: nil,
			Duration:     250 * time.Millisecond,
			Details:      map[string][]rule.LintError{},
			OrderedPaths: []string{},
		}

		var buf bytes.Buffer
		err := formatter.Format(&buf, result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, `"files": 2`) {
			t.Errorf("expected files count in JSON, got: %s", output)
		}
		if !strings.Contains(output, `"lines": 100`) {
			t.Errorf("expected lines count in JSON, got: %s", output)
		}
		if !strings.Contains(output, `"errors": 0`) {
			t.Errorf("expected errors count in JSON, got: %s", output)
		}
		if !strings.Contains(output, `"elapsed_ms": 250`) {
			t.Errorf("expected elapsed_ms in JSON, got: %s", output)
		}
		if strings.Contains(output, `"links_checked"`) {
			t.Errorf("should not include links_checked when nil, got: %s", output)
		}
	})

	t.Run("WithErrors", func(t *testing.T) {
		formatter := NewJSONFormatter()
		result := &Result{
			Files:  1,
			Lines:  50,
			Errors: 2,
			Details: map[string][]rule.LintError{
				"test.md": {
					{File: "test.md", Line: 10, Message: "Error 1"},
					{File: "test.md", Line: 20, Message: "Error 2"},
				},
			},
			Duration: 500 * time.Millisecond,
		}

		var buf bytes.Buffer
		err := formatter.Format(&buf, result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, `"errors": 2`) {
			t.Errorf("expected errors count, got: %s", output)
		}
		if !strings.Contains(output, `"test.md"`) {
			t.Errorf("expected filename in details, got: %s", output)
		}
		if !strings.Contains(output, `"Error 1"`) {
			t.Errorf("expected error message 1, got: %s", output)
		}
		if !strings.Contains(output, `"Error 2"`) {
			t.Errorf("expected error message 2, got: %s", output)
		}
	})

	t.Run("WithLinkCheck", func(t *testing.T) {
		formatter := NewJSONFormatter()
		linksChecked := 15
		result := &Result{
			Files:        3,
			Lines:        200,
			Errors:       0,
			LinksChecked: &linksChecked,
			Duration:     1500 * time.Millisecond,
			Details:      map[string][]rule.LintError{},
		}

		var buf bytes.Buffer
		err := formatter.Format(&buf, result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, `"links_checked": 15`) {
			t.Errorf("expected links_checked in JSON, got: %s", output)
		}
	})

	t.Run("ValidJSON", func(t *testing.T) {
		formatter := NewJSONFormatter()
		linksChecked := 10
		result := &Result{
			Files:        2,
			Lines:        100,
			Errors:       1,
			LinksChecked: &linksChecked,
			Duration:     300 * time.Millisecond,
			Details: map[string][]rule.LintError{
				"file.md": {
					{File: "file.md", Line: 5, Message: "Test error"},
				},
			},
			OrderedPaths: []string{"file.md"},
		}

		var buf bytes.Buffer
		err := formatter.Format(&buf, result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify it's valid JSON by unmarshaling
		var decoded map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
			t.Fatalf("output is not valid JSON: %v", err)
		}

		// Check expected fields exist
		if _, ok := decoded["files"]; !ok {
			t.Error("missing 'files' field in JSON output")
		}
		if _, ok := decoded["lines"]; !ok {
			t.Error("missing 'lines' field in JSON output")
		}
		if _, ok := decoded["errors"]; !ok {
			t.Error("missing 'errors' field in JSON output")
		}
		if _, ok := decoded["elapsed_ms"]; !ok {
			t.Error("missing 'elapsed_ms' field in JSON output")
		}
		if _, ok := decoded["details"]; !ok {
			t.Error("missing 'details' field in JSON output")
		}
		if _, ok := decoded["links_checked"]; !ok {
			t.Error("missing 'links_checked' field in JSON output")
		}
	})

	t.Run("WriteError", func(t *testing.T) {
		formatter := NewJSONFormatter()
		result := &Result{
			Files:    1,
			Lines:    10,
			Errors:   0,
			Duration: 100 * time.Millisecond,
			Details:  map[string][]rule.LintError{},
		}

		ew := &errorWriter{}
		err := formatter.Format(ew, result)
		if err == nil {
			t.Error("expected error when writing to errorWriter")
		}
	})
}

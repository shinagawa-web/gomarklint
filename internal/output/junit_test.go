package output

import (
	"bytes"
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/shinagawa-web/gomarklint/v3/internal/rule"
)

func TestJUnitFormatter_NoErrors(t *testing.T) {
	formatter := NewJUnitFormatter()
	result := &Result{
		Files:        2,
		Lines:        100,
		Total:        0,
		Duration:     420 * time.Millisecond,
		Details:      map[string][]rule.LintError{},
		OrderedPaths: []string{"README.md"},
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	assertContains(t, output, `<?xml version="1.0" encoding="UTF-8"?>`)
	assertContains(t, output, `name="gomarklint"`)
	assertContains(t, output, `failures="0"`)
	assertContains(t, output, `name="README.md"`)
	// passing file: single testcase with no failure child
	assertNotContains(t, output, "<failure")
}

func TestJUnitFormatter_WithErrors(t *testing.T) {
	formatter := NewJUnitFormatter()
	result := &Result{
		Files: 1,
		Lines: 50,
		Total: 2,
		Details: map[string][]rule.LintError{
			"docs/guide.md": {
				{File: "docs/guide.md", Line: 12, Message: "Link unreachable: https://example.com/broken", Severity: "error"},
				{File: "docs/guide.md", Line: 5, Message: "heading-increment", Severity: "error"},
			},
		},
		Duration:     420 * time.Millisecond,
		OrderedPaths: []string{"docs/guide.md"},
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	assertContains(t, output, `tests="2"`)
	assertContains(t, output, `failures="2"`)
	assertContains(t, output, `name="docs/guide.md"`)
	assertContains(t, output, `line 12: Link unreachable`)
	assertContains(t, output, `line 5: heading-increment`)
	assertContains(t, output, `type="error"`)
}

func TestJUnitFormatter_MixedFiles(t *testing.T) {
	formatter := NewJUnitFormatter()
	result := &Result{
		Files: 2,
		Lines: 80,
		Total: 1,
		Details: map[string][]rule.LintError{
			"docs/guide.md": {
				{File: "docs/guide.md", Line: 12, Message: "Link unreachable", Severity: "error"},
			},
		},
		Duration:     420 * time.Millisecond,
		OrderedPaths: []string{"docs/guide.md", "README.md"},
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// total tests: 1 violation + 1 passing testcase = 2
	assertContains(t, output, `tests="2"`)
	assertContains(t, output, `failures="1"`)
}

func TestJUnitFormatter_WarningType(t *testing.T) {
	formatter := NewJUnitFormatter()
	result := &Result{
		Files: 1,
		Lines: 20,
		Total: 1,
		Details: map[string][]rule.LintError{
			"file.md": {
				{File: "file.md", Line: 3, Message: "setext heading", Severity: "warning"},
			},
		},
		Duration:     100 * time.Millisecond,
		OrderedPaths: []string{"file.md"},
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertContains(t, buf.String(), `type="warning"`)
}

func TestJUnitFormatter_DefaultSeverity(t *testing.T) {
	formatter := NewJUnitFormatter()
	result := &Result{
		Files: 1,
		Lines: 10,
		Total: 1,
		Details: map[string][]rule.LintError{
			"file.md": {
				{File: "file.md", Line: 1, Message: "some issue"},
			},
		},
		Duration:     50 * time.Millisecond,
		OrderedPaths: []string{"file.md"},
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertContains(t, buf.String(), `type="error"`)
}

func TestJUnitFormatter_TimeFormatting(t *testing.T) {
	formatter := NewJUnitFormatter()
	result := &Result{
		Files:        1,
		Lines:        10,
		Total:        0,
		Duration:     1234 * time.Millisecond,
		Details:      map[string][]rule.LintError{},
		OrderedPaths: []string{"file.md"},
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), `time="1.23"`) {
		t.Errorf("expected time=1.23, got: %s", buf.String())
	}
}

func TestJUnitFormatter_ValidXML(t *testing.T) {
	formatter := NewJUnitFormatter()
	result := &Result{
		Files: 2,
		Lines: 100,
		Total: 1,
		Details: map[string][]rule.LintError{
			"docs/guide.md": {
				{File: "docs/guide.md", Line: 12, Message: "Link unreachable: https://example.com/broken", Severity: "error"},
			},
		},
		Duration:     420 * time.Millisecond,
		OrderedPaths: []string{"docs/guide.md", "README.md"},
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded interface{}
	if err := xml.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid XML: %v", err)
	}
}

func TestJUnitFormatter_WriteError(t *testing.T) {
	formatter := NewJUnitFormatter()
	result := &Result{
		Files:        1,
		Lines:        10,
		Total:        0,
		Duration:     100 * time.Millisecond,
		Details:      map[string][]rule.LintError{},
		OrderedPaths: []string{},
	}

	ew := &errorWriter{}
	err := formatter.Format(ew, result)
	if err == nil {
		t.Error("expected error when writing to errorWriter")
	}
}

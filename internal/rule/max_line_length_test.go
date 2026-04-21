package rule

import (
	"fmt"
	"strings"
	"testing"
)

func TestCheckMaxLineLength(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		lineLength int
		offset     int
		wantErrs   []LintError
	}{
		// ── valid cases ──────────────────────────────────────────────────────
		{
			name:       "valid: short line",
			content:    "Short line.\n",
			lineLength: 80,
			wantErrs:   nil,
		},
		{
			name:       "valid: line exactly at limit",
			content:    strings.Repeat("a", 80) + "\n",
			lineLength: 80,
			wantErrs:   nil,
		},
		{
			name:       "valid: heading exempt",
			content:    "## " + strings.Repeat("a", 100) + "\n",
			lineLength: 80,
			wantErrs:   nil,
		},
		{
			name:       "valid: h1 heading exempt",
			content:    "# " + strings.Repeat("a", 100) + "\n",
			lineLength: 80,
			wantErrs:   nil,
		},
		{
			name:       "valid: bare https URL line exempt",
			content:    "https://example.com/very/long/path/that/exceeds/eighty/characters/in/total/length\n",
			lineLength: 80,
			wantErrs:   nil,
		},
		{
			name:       "valid: bare http URL line exempt",
			content:    "http://example.com/very/long/path/that/exceeds/eighty/characters/in/total/length\n",
			lineLength: 80,
			wantErrs:   nil,
		},
		{
			name:       "valid: angle-bracket URL line exempt",
			content:    "<https://example.com/very/long/path/that/exceeds/eighty/characters/in/total/>\n",
			lineLength: 80,
			wantErrs:   nil,
		},
		{
			name:       "valid: long line inside fenced code block",
			content:    "```go\n" + strings.Repeat("x", 100) + "\n```\n",
			lineLength: 80,
			wantErrs:   nil,
		},
		{
			name:       "valid: long line inside tilde fence",
			content:    "~~~sh\n" + strings.Repeat("y", 100) + "\n~~~\n",
			lineLength: 80,
			wantErrs:   nil,
		},
		{
			name:       "valid: empty file",
			content:    "",
			lineLength: 80,
			wantErrs:   nil,
		},
		// ── invalid cases ────────────────────────────────────────────────────
		{
			name:       "invalid: line one char over limit",
			content:    strings.Repeat("a", 81) + "\n",
			lineLength: 80,
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "max-line-length: line exceeds 80 characters (81)"},
			},
		},
		{
			name:       "invalid: multiple violations",
			content:    strings.Repeat("a", 85) + "\n" + "ok\n" + strings.Repeat("b", 90) + "\n",
			lineLength: 80,
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "max-line-length: line exceeds 80 characters (85)"},
				{File: "test.md", Line: 3, Message: "max-line-length: line exceeds 80 characters (90)"},
			},
		},
		{
			name:       "invalid: custom line length",
			content:    strings.Repeat("a", 101) + "\n",
			lineLength: 100,
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "max-line-length: line exceeds 100 characters (101)"},
			},
		},
		{
			name:       "invalid: offset shifts line numbers",
			content:    strings.Repeat("a", 81) + "\n",
			lineLength: 80,
			offset:     5,
			wantErrs: []LintError{
				{File: "test.md", Line: 6, Message: "max-line-length: line exceeds 80 characters (81)"},
			},
		},
		{
			name: "invalid: prose line with URL in middle",
			// "See https://example.com for details on " (39) + 50 "a"s + "." = 90 chars
			content:    "See https://example.com for details on " + strings.Repeat("a", 50) + ".\n",
			lineLength: 80,
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: fmt.Sprintf("max-line-length: line exceeds 80 characters (%d)", len("See https://example.com for details on "+strings.Repeat("a", 50)+"."))},
			},
		},
		{
			name:       "valid: long line after closing fence is checked",
			content:    "```go\n" + strings.Repeat("x", 100) + "\n```\n" + strings.Repeat("y", 81) + "\n",
			lineLength: 80,
			wantErrs: []LintError{
				{File: "test.md", Line: 4, Message: "max-line-length: line exceeds 80 characters (81)"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(tt.content, "\n")
			got := CheckMaxLineLength("test.md", lines, tt.offset, tt.lineLength)

			if len(got) != len(tt.wantErrs) {
				t.Fatalf("got %d errors, want %d\ngot:  %v\nwant: %v", len(got), len(tt.wantErrs), got, tt.wantErrs)
			}
			for i := range got {
				if got[i].File != tt.wantErrs[i].File {
					t.Errorf("[%d] file: got %q, want %q", i, got[i].File, tt.wantErrs[i].File)
				}
				if got[i].Line != tt.wantErrs[i].Line {
					t.Errorf("[%d] line: got %d, want %d", i, got[i].Line, tt.wantErrs[i].Line)
				}
				if got[i].Message != tt.wantErrs[i].Message {
					t.Errorf("[%d] message: got %q, want %q", i, got[i].Message, tt.wantErrs[i].Message)
				}
			}
		})
	}
}

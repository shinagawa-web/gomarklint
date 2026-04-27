package rule

import (
	"strings"
	"testing"
)

func TestCheckBlanksAroundFences(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		offset   int
		wantErrs []LintError
	}{
		{
			name:     "valid: blank lines around fence",
			content:  "Some text\n\n```go\ncode\n```\n\nMore text\n",
			wantErrs: nil,
		},
		{
			name:     "valid: fence at start of file needs no preceding blank",
			content:  "```go\ncode\n```\n\nMore text\n",
			wantErrs: nil,
		},
		{
			name:     "valid: fence at end of file needs no following blank",
			content:  "Some text\n\n```go\ncode\n```",
			wantErrs: nil,
		},
		{
			name:     "valid: fence is the only content",
			content:  "```go\ncode\n```",
			wantErrs: nil,
		},
		{
			name:     "valid: tilde fence with blank lines",
			content:  "Some text\n\n~~~go\ncode\n~~~\n\nMore text\n",
			wantErrs: nil,
		},
		{
			name:    "invalid: no blank line before opening fence",
			content: "Some text\n```go\ncode\n```\n\nMore text\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "blanks-around-fences: fenced code block must be preceded by a blank line"},
			},
		},
		{
			name:    "invalid: no blank line after closing fence",
			content: "Some text\n\n```go\ncode\n```\nMore text\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 5, Message: "blanks-around-fences: fenced code block must be followed by a blank line"},
			},
		},
		{
			name:    "invalid: missing blank before and after",
			content: "Some text\n```go\ncode\n```\nMore text\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "blanks-around-fences: fenced code block must be preceded by a blank line"},
				{File: "test.md", Line: 4, Message: "blanks-around-fences: fenced code block must be followed by a blank line"},
			},
		},
		{
			name:    "invalid: tilde fence missing blank before",
			content: "Some text\n~~~go\ncode\n~~~\n\nMore text\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "blanks-around-fences: fenced code block must be preceded by a blank line"},
			},
		},
		{
			name:    "invalid: multiple fences with violations",
			content: "text\n```\nfoo\n```\ntext\n```\nbar\n```\nend\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "blanks-around-fences: fenced code block must be preceded by a blank line"},
				{File: "test.md", Line: 4, Message: "blanks-around-fences: fenced code block must be followed by a blank line"},
				{File: "test.md", Line: 6, Message: "blanks-around-fences: fenced code block must be preceded by a blank line"},
				{File: "test.md", Line: 8, Message: "blanks-around-fences: fenced code block must be followed by a blank line"},
			},
		},
		{
			name:    "offset shifts line numbers",
			content: "Some text\n```go\ncode\n```\n\nMore text\n",
			offset:  5,
			wantErrs: []LintError{
				{File: "test.md", Line: 7, Message: "blanks-around-fences: fenced code block must be preceded by a blank line"},
			},
		},
		{
			name:     "empty file",
			content:  "",
			wantErrs: nil,
		},
		{
			name:     "valid: longer fence marker",
			content:  "Some text\n\n````go\ncode\n````\n\nMore text\n",
			wantErrs: nil,
		},
		{
			name:     "valid: fence inside HTML comment block is ignored",
			content:  "Some text\n<!--\n```go\ncode\n```\n-->\nMore text\n",
			wantErrs: nil,
		},
		{
			name:     "valid: fence after HTML comment block requires blank line",
			content:  "<!--\ncomment\n-->\n\n```go\ncode\n```\n\nMore text\n",
			wantErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(tt.content, "\n")
			got := CheckBlanksAroundFences("test.md", lines, tt.offset)

			if len(got) != len(tt.wantErrs) {
				t.Fatalf("got %d errors, want %d\ngot:  %v\nwant: %v", len(got), len(tt.wantErrs), got, tt.wantErrs)
			}
			for i := range got {
				if got[i] != tt.wantErrs[i] {
					t.Errorf("error[%d]: got %+v, want %+v", i, got[i], tt.wantErrs[i])
				}
			}
		})
	}
}

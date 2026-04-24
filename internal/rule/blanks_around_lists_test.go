package rule

import (
	"strings"
	"testing"
)

func TestCheckBlanksAroundLists(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		offset   int
		wantErrs []LintError
	}{
		{
			name:     "valid: unordered list surrounded by blank lines",
			content:  "Some text\n\n- item 1\n- item 2\n\nMore text\n",
			wantErrs: nil,
		},
		{
			name:     "valid: ordered list surrounded by blank lines",
			content:  "Some text\n\n1. First\n2. Second\n\nMore text\n",
			wantErrs: nil,
		},
		{
			name:     "valid: list at start of file needs no preceding blank",
			content:  "- item 1\n- item 2\n\nSome text\n",
			wantErrs: nil,
		},
		{
			name:     "valid: list at end of file needs no following blank",
			content:  "Some text\n\n- item 1\n- item 2\n",
			wantErrs: nil,
		},
		{
			name:     "valid: list is the only content",
			content:  "- item 1\n- item 2\n",
			wantErrs: nil,
		},
		{
			name:     "valid: nested list requires no blank between parent and child",
			content:  "Some text\n\n- item 1\n  - nested 1\n  - nested 2\n- item 2\n\nMore text\n",
			wantErrs: nil,
		},
		{
			name:     "valid: list inside fenced code block is ignored",
			content:  "Some text\n\n```\n- not a list\n```\n\nMore text\n",
			wantErrs: nil,
		},
		{
			name:     "valid: list inside tilde fence is ignored",
			content:  "Some text\n\n~~~\n- not a list\n~~~\n\nMore text\n",
			wantErrs: nil,
		},
		{
			name:     "valid: ordered list with parenthesis marker",
			content:  "Some text\n\n1) First\n2) Second\n\nMore text\n",
			wantErrs: nil,
		},
		{
			name:    "invalid: missing blank line before unordered list",
			content: "Some text\n- item 1\n- item 2\n\nMore text\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "blanks-around-lists: list must be preceded by a blank line"},
			},
		},
		{
			name:    "invalid: missing blank line after unordered list",
			content: "Some text\n\n- item 1\n- item 2\nMore text\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 5, Message: "blanks-around-lists: list must be followed by a blank line"},
			},
		},
		{
			name:    "invalid: missing blank lines before and after list",
			content: "Some text\n- item 1\n- item 2\nMore text\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "blanks-around-lists: list must be preceded by a blank line"},
				{File: "test.md", Line: 4, Message: "blanks-around-lists: list must be followed by a blank line"},
			},
		},
		{
			name:    "invalid: missing blank line before ordered list",
			content: "Some text\n1. First\n2. Second\n\nMore text\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "blanks-around-lists: list must be preceded by a blank line"},
			},
		},
		{
			name:    "invalid: missing blank line after ordered list",
			content: "Some text\n\n1. First\n2. Second\nMore text\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 5, Message: "blanks-around-lists: list must be followed by a blank line"},
			},
		},
		{
			name:    "invalid: multiple list blocks with violations",
			content: "Some text\n- item 1\n\nOther text\n- item 2\nEnd text\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "blanks-around-lists: list must be preceded by a blank line"},
				{File: "test.md", Line: 5, Message: "blanks-around-lists: list must be preceded by a blank line"},
				{File: "test.md", Line: 6, Message: "blanks-around-lists: list must be followed by a blank line"},
			},
		},
		{
			name:    "invalid: list after fenced code block without blank line",
			content: "```go\nfmt.Println()\n```\n- item 1\n\nText\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 4, Message: "blanks-around-lists: list must be preceded by a blank line"},
			},
		},
		{
			name:    "invalid: list immediately followed by fenced code block opener",
			content: "- item 1\n- item 2\n```go\nfmt.Println()\n```\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "blanks-around-lists: list must be followed by a blank line"},
			},
		},
		{
			name:    "invalid: list immediately followed by tilde fence opener",
			content: "- item 1\n- item 2\n~~~go\nfmt.Println()\n~~~\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "blanks-around-lists: list must be followed by a blank line"},
			},
		},
		{
			name:    "offset shifts line numbers",
			content: "Some text\n- item 1\nMore text\n",
			offset:  5,
			wantErrs: []LintError{
				{File: "test.md", Line: 7, Message: "blanks-around-lists: list must be preceded by a blank line"},
				{File: "test.md", Line: 8, Message: "blanks-around-lists: list must be followed by a blank line"},
			},
		},
		{
			name:     "empty file",
			content:  "",
			wantErrs: nil,
		},
		{
			name:     "valid: star and plus unordered markers",
			content:  "Some text\n\n* item 1\n+ item 2\n\nMore text\n",
			wantErrs: nil,
		},
		{
			name:    "invalid: star marker missing blank before",
			content: "Some text\n* item 1\n\nMore text\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "blanks-around-lists: list must be preceded by a blank line"},
			},
		},
		{
			name:     "valid: unordered marker followed by tab",
			content:  "Some text\n\n-\titem 1\n\nMore text\n",
			wantErrs: nil,
		},
		{
			name:     "valid: ordered marker followed by tab",
			content:  "Some text\n\n1.\tFirst\n\nMore text\n",
			wantErrs: nil,
		},
		{
			name:    "invalid: tab-separated unordered list missing blank before",
			content: "Some text\n-\titem 1\n\nMore text\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "blanks-around-lists: list must be preceded by a blank line"},
			},
		},
		{
			name:    "invalid: tab-separated ordered list missing blank before",
			content: "Some text\n1.\tFirst\n\nMore text\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "blanks-around-lists: list must be preceded by a blank line"},
			},
		},
		{
			name:     "valid: digits followed by non-separator char are not a list item",
			content:  "Some text\n12x foo\nMore text\n",
			wantErrs: nil,
		},
		{
			name:     "valid: ordered marker at end of line is not a list item",
			content:  "Some text\n12.\nMore text\n",
			wantErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(tt.content, "\n")
			got := CheckBlanksAroundLists("test.md", lines, tt.offset)

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

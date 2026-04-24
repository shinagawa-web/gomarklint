package rule

import (
	"strings"
	"testing"
)

func TestCheckNoHardTabs(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		offset   int
		wantErrs []LintError
	}{
		{
			name:     "valid: no tabs",
			content:  "- item\n  - nested\n",
			wantErrs: nil,
		},
		{
			name:    "invalid: tab at start of line",
			content: "\t- item\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-hard-tabs: hard tab character found at column 1"},
			},
		},
		{
			name:    "invalid: tab in paragraph",
			content: "key\tvalue\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-hard-tabs: hard tab character found at column 4"},
			},
		},
		{
			name:    "invalid: multiple tabs on one line",
			content: "\tcol1\tcol2\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-hard-tabs: hard tab character found at column 1"},
				{File: "test.md", Line: 1, Message: "no-hard-tabs: hard tab character found at column 6"},
			},
		},
		{
			name:     "valid: tab inside fenced code block (backtick)",
			content:  "```\n\tcode\n```\n",
			wantErrs: nil,
		},
		{
			name:     "valid: tab inside fenced code block (tilde)",
			content:  "~~~\n\tcode\n~~~\n",
			wantErrs: nil,
		},
		{
			name:     "valid: tab inside inline code span",
			content:  "Use `key\tvalue` as example.\n",
			wantErrs: nil,
		},
		{
			name:     "valid: tab inside double-backtick inline code",
			content:  "Use ``key\tvalue`` as example.\n",
			wantErrs: nil,
		},
		{
			name:    "invalid: tab outside inline code on same line",
			content: "Run `cmd` then\there.\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-hard-tabs: hard tab character found at column 15"},
			},
		},
		{
			name:    "invalid: tab on second line",
			content: "# Heading\n\n\tindented\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "no-hard-tabs: hard tab character found at column 1"},
			},
		},
		{
			name:    "offset shifts line numbers",
			content: "\titem\n",
			offset:  5,
			wantErrs: []LintError{
				{File: "test.md", Line: 6, Message: "no-hard-tabs: hard tab character found at column 1"},
			},
		},
		{
			name:     "valid: empty file",
			content:  "",
			wantErrs: nil,
		},
		{
			name:    "invalid: tab before and after fenced code block",
			content: "\tbefore\n```\n\tcode\n```\n\tafter\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-hard-tabs: hard tab character found at column 1"},
				{File: "test.md", Line: 5, Message: "no-hard-tabs: hard tab character found at column 1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(tt.content, "\n")
			got := CheckNoHardTabs("test.md", lines, tt.offset)

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

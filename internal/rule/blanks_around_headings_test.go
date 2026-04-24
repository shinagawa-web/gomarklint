package rule

import (
	"strings"
	"testing"
)

func TestCheckBlanksAroundHeadings(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		offset   int
		wantErrs []LintError
	}{
		{
			name:     "valid: blank lines around all headings",
			content:  "## Introduction\n\nSome text\n\n## Section\n\nMore text\n",
			wantErrs: nil,
		},
		{
			name:     "valid: heading at start of file needs no preceding blank",
			content:  "## Title\n\nSome text\n",
			wantErrs: nil,
		},
		{
			name:     "valid: heading at end of file needs no following blank",
			content:  "Some text\n\n## Title",
			wantErrs: nil,
		},
		{
			name:     "valid: heading is the only line",
			content:  "## Title",
			wantErrs: nil,
		},
		{
			name:     "valid: ATX heading with tab after hash run",
			content:  "##\tTitle\n\nSome text\n",
			wantErrs: nil,
		},
		{
			name:    "invalid: ATX heading with tab, no blank line after",
			content: "Some text\n\n##\tTitle\nMore text\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "blanks-around-headings: heading must be followed by a blank line"},
			},
		},
		{
			name:     "valid: all heading levels with blank lines",
			content:  "# H1\n\n## H2\n\n### H3\n\n#### H4\n\n##### H5\n\n###### H6\n",
			wantErrs: nil,
		},
		{
			name:    "invalid: no blank line before heading",
			content: "Some text\n## Heading\n\nMore text\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "blanks-around-headings: heading must be preceded by a blank line"},
			},
		},
		{
			name:    "invalid: no blank line after heading",
			content: "Some text\n\n## Heading\nMore text\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "blanks-around-headings: heading must be followed by a blank line"},
			},
		},
		{
			name:    "invalid: no blank lines before or after",
			content: "Some text\n## Heading\nMore text\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "blanks-around-headings: heading must be preceded by a blank line"},
				{File: "test.md", Line: 2, Message: "blanks-around-headings: heading must be followed by a blank line"},
			},
		},
		{
			name:    "invalid: multiple violations across headings",
			content: "Some text\n## Section One\nMore text\n\n## Section Two\nEven more\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "blanks-around-headings: heading must be preceded by a blank line"},
				{File: "test.md", Line: 2, Message: "blanks-around-headings: heading must be followed by a blank line"},
				{File: "test.md", Line: 5, Message: "blanks-around-headings: heading must be followed by a blank line"},
			},
		},
		{
			name:     "valid: heading inside fenced code block is ignored",
			content:  "Some text\n\n```\n## Not a heading\n```\n\nMore text\n",
			wantErrs: nil,
		},
		{
			name:     "valid: heading inside tilde fence is ignored",
			content:  "Some text\n\n~~~\n## Not a heading\n~~~\n\nMore text\n",
			wantErrs: nil,
		},
		{
			name:    "invalid: heading after fenced block without blank line",
			content: "```go\nfmt.Println()\n```\n## Heading\n\nText\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 4, Message: "blanks-around-headings: heading must be preceded by a blank line"},
			},
		},
		{
			name:    "invalid: heading immediately followed by fenced code block opener",
			content: "## Heading\n```go\nfmt.Println()\n```\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "blanks-around-headings: heading must be followed by a blank line"},
			},
		},
		{
			name:    "invalid: heading immediately followed by tilde fence opener",
			content: "## Heading\n~~~go\nfmt.Println()\n~~~\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "blanks-around-headings: heading must be followed by a blank line"},
			},
		},
		{
			name:    "offset shifts line numbers",
			content: "Some text\n## Heading\nMore text\n",
			offset:  5,
			wantErrs: []LintError{
				{File: "test.md", Line: 7, Message: "blanks-around-headings: heading must be preceded by a blank line"},
				{File: "test.md", Line: 7, Message: "blanks-around-headings: heading must be followed by a blank line"},
			},
		},
		{
			name:     "empty file",
			content:  "",
			wantErrs: nil,
		},
		{
			name:     "non-heading hash sequences are ignored",
			content:  "Some text\n#notaheading\nMore text\n",
			wantErrs: nil,
		},
		{
			name:     "seven hashes is not a heading",
			content:  "Some text\n####### Too deep\nMore text\n",
			wantErrs: nil,
		},
		{
			name:     "valid: consecutive headings with blank lines between",
			content:  "## First\n\n## Second\n\n## Third\n",
			wantErrs: nil,
		},
		{
			name:    "invalid: consecutive headings without blank lines",
			content: "## First\n## Second\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "blanks-around-headings: heading must be followed by a blank line"},
				{File: "test.md", Line: 2, Message: "blanks-around-headings: heading must be preceded by a blank line"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(tt.content, "\n")
			got := CheckBlanksAroundHeadings("test.md", lines, tt.offset)

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

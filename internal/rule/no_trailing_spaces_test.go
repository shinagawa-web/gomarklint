package rule

import (
	"strings"
	"testing"
)

func TestCheckNoTrailingSpaces(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		offset   int
		wantErrs []LintError
	}{
		{
			name:     "valid: no trailing whitespace",
			content:  "This is a clean line\n",
			wantErrs: nil,
		},
		{
			name:     "valid: CRLF line endings without trailing spaces",
			content:  "Clean line\r\nAnother clean line\r\n",
			wantErrs: nil,
		},
		{
			name:    "invalid: trailing space on CRLF line",
			content: "This line has trailing spaces   \r\nClean line\r\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-trailing-spaces: trailing whitespace found"},
			},
		},
		{
			name:    "invalid: trailing tab on CRLF line",
			content: "This line has a trailing tab\t\r\nClean line\r\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-trailing-spaces: trailing whitespace found"},
			},
		},
		{
			name:     "valid: empty file",
			content:  "",
			wantErrs: nil,
		},
		{
			name:     "valid: blank line",
			content:  "\n",
			wantErrs: nil,
		},
		{
			name:    "invalid: single trailing space",
			content: "This line has a trailing space \n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-trailing-spaces: trailing whitespace found"},
			},
		},
		{
			name:    "invalid: multiple trailing spaces",
			content: "This line has trailing spaces   \n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-trailing-spaces: trailing whitespace found"},
			},
		},
		{
			name:    "invalid: trailing tab",
			content: "This line has a trailing tab\t\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-trailing-spaces: trailing whitespace found"},
			},
		},
		{
			name:    "invalid: mixed trailing whitespace",
			content: "Trailing mix \t\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-trailing-spaces: trailing whitespace found"},
			},
		},
		{
			name:    "invalid: multiple lines with violations",
			content: "Line one \nLine two\nLine three \n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-trailing-spaces: trailing whitespace found"},
				{File: "test.md", Line: 3, Message: "no-trailing-spaces: trailing whitespace found"},
			},
		},
		{
			name:     "valid: trailing space inside fenced code block is ignored",
			content:  "```\ncode line   \n```\n",
			wantErrs: nil,
		},
		{
			name:     "valid: trailing space inside tilde fence is ignored",
			content:  "~~~\ncode line   \n~~~\n",
			wantErrs: nil,
		},
		{
			name:    "invalid: trailing space before fenced block",
			content: "Text before \n```\ncode\n```\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-trailing-spaces: trailing whitespace found"},
			},
		},
		{
			name:    "invalid: trailing space after fenced block",
			content: "```\ncode\n```\nText after \n",
			wantErrs: []LintError{
				{File: "test.md", Line: 4, Message: "no-trailing-spaces: trailing whitespace found"},
			},
		},
		{
			name:    "offset shifts line numbers",
			content: "trailing space \n",
			offset:  5,
			wantErrs: []LintError{
				{File: "test.md", Line: 6, Message: "no-trailing-spaces: trailing whitespace found"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(tt.content, "\n")
			got := CheckNoTrailingSpaces("test.md", lines, tt.offset)

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

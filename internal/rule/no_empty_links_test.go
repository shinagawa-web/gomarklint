package rule

import (
	"strings"
	"testing"
)

func TestCheckNoEmptyLinks(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		offset   int
		wantErrs []LintError
	}{
		{
			name:     "valid: normal link",
			content:  "[Example](https://example.com)\n",
			wantErrs: nil,
		},
		{
			name:     "valid: fragment link",
			content:  "[Section](#section)\n",
			wantErrs: nil,
		},
		{
			name:     "valid: relative path",
			content:  "[Page](./page.md)\n",
			wantErrs: nil,
		},
		{
			name:     "valid: image with URL",
			content:  "![Alt](https://example.com/img.png)\n",
			wantErrs: nil,
		},
		{
			name:    "invalid: empty link destination",
			content: "[click here]()\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-empty-links: link has empty destination: [click here]()"},
			},
		},
		{
			name:    "invalid: fragment-only destination",
			content: "[click here](#)\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-empty-links: link has empty destination: [click here](#)"},
			},
		},
		{
			name:    "invalid: angle bracket empty destination",
			content: "[click here](<>)\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-empty-links: link has empty destination: [click here](<>)"},
			},
		},
		{
			name:    "invalid: empty image destination",
			content: "![alt text]()\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-empty-links: link has empty destination: ![alt text]()"},
			},
		},
		{
			name:    "invalid: image with fragment-only destination",
			content: "![alt](#)\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-empty-links: link has empty destination: ![alt](#)"},
			},
		},
		{
			name:     "valid: link inside fenced code block",
			content:  "```\n[text]()\n```\n",
			wantErrs: nil,
		},
		{
			name:     "valid: link inside tilde fence",
			content:  "~~~\n[text]()\n~~~\n",
			wantErrs: nil,
		},
		{
			name:     "valid: link inside inline code",
			content:  "Use `[text]()` as example.\n",
			wantErrs: nil,
		},
		{
			name:     "valid: link inside double-backtick inline code",
			content:  "Use ``[text]()`` as example.\n",
			wantErrs: nil,
		},
		{
			name:    "invalid: multiple empty links on one line",
			content: "[a]() and [b](#)\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-empty-links: link has empty destination: [a]()"},
				{File: "test.md", Line: 1, Message: "no-empty-links: link has empty destination: [b](#)"},
			},
		},
		{
			name:    "invalid: empty link on second line",
			content: "# Heading\n\n[broken]()\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "no-empty-links: link has empty destination: [broken]()"},
			},
		},
		{
			name:    "offset shifts line numbers",
			content: "[text]()\n",
			offset:  5,
			wantErrs: []LintError{
				{File: "test.md", Line: 6, Message: "no-empty-links: link has empty destination: [text]()"},
			},
		},
		{
			name:     "valid: no links at all",
			content:  "Just some plain text.\n",
			wantErrs: nil,
		},
		{
			name:     "empty file",
			content:  "",
			wantErrs: nil,
		},
		{
			name:    "invalid: empty link after inline code on same line",
			content: "Run `cmd` then [see]()\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-empty-links: link has empty destination: [see]()"},
			},
		},
		{
			name:     "valid: unclosed paren is not a link",
			content:  "[text](no closing paren\n",
			wantErrs: nil,
		},
		{
			name:    "valid: destination with only spaces is empty",
			content: "[text](   )\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-empty-links: link has empty destination: [text](   )"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(tt.content, "\n")
			got := CheckNoEmptyLinks("test.md", lines, tt.offset)

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

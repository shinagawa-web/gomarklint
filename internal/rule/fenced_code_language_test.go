package rule

import (
	"strings"
	"testing"
)

func TestCheckFencedCodeLanguage(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		offset   int
		wantErrs []LintError
	}{
		{
			name:     "no code blocks",
			content:  "# Hello\n\nSome text.\n",
			wantErrs: nil,
		},
		{
			name:     "code block with language",
			content:  "# Hello\n\n```go\nfmt.Println()\n```\n",
			wantErrs: nil,
		},
		{
			name:    "code block without language",
			content: "# Hello\n\n```\nsome code\n```\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "Fenced code block must have a language identifier"},
			},
		},
		{
			name:     "tilde fence with language",
			content:  "~~~python\nprint('hi')\n~~~\n",
			wantErrs: nil,
		},
		{
			name:    "tilde fence without language",
			content: "~~~\nsome code\n~~~\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "Fenced code block must have a language identifier"},
			},
		},
		{
			name:    "multiple blocks, one without language",
			content: "```go\nfmt.Println()\n```\n\n```\nno lang\n```\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 5, Message: "Fenced code block must have a language identifier"},
			},
		},
		{
			name:    "multiple blocks, all without language",
			content: "```\nfirst\n```\n\n```\nsecond\n```\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "Fenced code block must have a language identifier"},
				{File: "test.md", Line: 5, Message: "Fenced code block must have a language identifier"},
			},
		},
		{
			name:     "closing fence is not flagged",
			content:  "```go\ncode\n```\n",
			wantErrs: nil,
		},
		{
			name:     "language with extra spaces after fence",
			content:  "``` go\ncode\n```\n",
			wantErrs: nil,
		},
		{
			name:     "empty file",
			content:  "",
			wantErrs: nil,
		},
		{
			name:    "offset shifts line numbers",
			content: "```\ncode\n```\n",
			offset:  5,
			wantErrs: []LintError{
				{File: "test.md", Line: 6, Message: "Fenced code block must have a language identifier"},
			},
		},
		{
			name:     "backtick fence inside tilde fence not treated as opening",
			content:  "~~~markdown\n```\ninner\n```\n~~~\n",
			wantErrs: nil,
		},
		{
			name:     "4-backtick fence with language",
			content:  "````go\ncode\n````\n",
			wantErrs: nil,
		},
		{
			name:    "4-backtick fence without language",
			content: "````\ncode\n````\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "Fenced code block must have a language identifier"},
			},
		},
		{
			name:     "4-tilde fence with language",
			content:  "~~~~python\ncode\n~~~~\n",
			wantErrs: nil,
		},
		{
			name:    "4-tilde fence without language",
			content: "~~~~\ncode\n~~~~\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "Fenced code block must have a language identifier"},
			},
		},
		{
			name:     "longer closing fence closes block",
			content:  "```go\ncode\n`````\n",
			wantErrs: nil,
		},
		{
			name:    "longer closing fence allows next block to be detected",
			content: "```go\ncode\n`````\n\n```\nno lang\n```\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 5, Message: "Fenced code block must have a language identifier"},
			},
		},
		{
			name:     "longer closing tilde fence",
			content:  "~~~py\ncode\n~~~~~\n",
			wantErrs: nil,
		},
		{
			name:     "fence opener inside single-line HTML comment is ignored",
			content:  "<!-- ```\nsome text\n",
			wantErrs: nil,
		},
		{
			name:     "fence opener inside multi-line HTML comment is ignored",
			content:  "<!--\n```\ncode\n```\n-->\n",
			wantErrs: nil,
		},
		{
			name:    "fence after closed multi-line HTML comment is flagged",
			content: "<!--\ncomment\n-->\n```\ncode\n```\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 4, Message: "Fenced code block must have a language identifier"},
			},
		},
		{
			name:     "fence opener on same line as --> closing comment is not flagged",
			content:  "<!-- comment --> ``` not a fence\ntext\n",
			wantErrs: nil,
		},
		{
			name:     "line starting with backtick but not a fence marker is skipped",
			content:  "`inline code` on its own line\ntext\n",
			wantErrs: nil,
		},
		{
			name:     "non-fence-starting line inside code block does not close block",
			content:  "```go\nsome code line\n```\n",
			wantErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(tt.content, "\n")
			got := CheckFencedCodeLanguage("test.md", lines, tt.offset)

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

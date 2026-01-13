package rule

import (
	"testing"
)

func TestCheckUnclosedCodeBlocks(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantErrs []LintError
	}{
		{
			name:     "valid closed code blocks",
			content:  "```\ncode\n```\n",
			wantErrs: nil,
		},
		{
			name:    "unclosed code block",
			content: "```\ncode\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "Unclosed code block"},
			},
		},
		{
			name:    "unclosed code block with language",
			content: "```go\nfunc main() {}\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "Unclosed code block"},
			},
		},
		{
			name:     "multiple valid code blocks",
			content:  "```\ncode1\n```\n\n```js\ncode2\n```\n",
			wantErrs: nil,
		},
		{
			name:     "no code blocks",
			content:  "# Just a heading\nSome text",
			wantErrs: nil,
		},
		{
			name:    "unclosed with frontmatter",
			content: "---\ntitle: Test\n---\n\n## Heading\n\n```\ncode\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 7, Message: "Unclosed code block"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckUnclosedCodeBlocks("test.md", tt.content)

			if len(got) != len(tt.wantErrs) {
				t.Fatalf("got %d errors, want %d\nGot: %v\nWant: %v", len(got), len(tt.wantErrs), got, tt.wantErrs)
			}

			for i := range got {
				if got[i].File != tt.wantErrs[i].File {
					t.Errorf("error %d: got file %q, want %q", i, got[i].File, tt.wantErrs[i].File)
				}
				if got[i].Line != tt.wantErrs[i].Line {
					t.Errorf("error %d: got line %d, want %d", i, got[i].Line, tt.wantErrs[i].Line)
				}
				if got[i].Message != tt.wantErrs[i].Message {
					t.Errorf("error %d: got message %q, want %q", i, got[i].Message, tt.wantErrs[i].Message)
				}
			}
		})
	}
}

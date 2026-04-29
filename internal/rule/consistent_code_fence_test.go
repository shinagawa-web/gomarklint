package rule

import (
	"strings"
	"testing"
)

func TestCheckConsistentCodeFence(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		offset   int
		style    string
		wantErrs []LintError
	}{
		// empty / no fences
		{
			name:     "empty doc",
			content:  "",
			style:    "consistent",
			wantErrs: nil,
		},
		{
			name:     "no fences",
			content:  "# Hello\n\nSome paragraph.\n",
			style:    "consistent",
			wantErrs: nil,
		},

		// single fence — never flagged in any mode
		{
			name:     "single backtick fence consistent",
			content:  "```go\ncode\n```\n",
			style:    "consistent",
			wantErrs: nil,
		},
		{
			name:     "single tilde fence consistent",
			content:  "~~~go\ncode\n~~~\n",
			style:    "consistent",
			wantErrs: nil,
		},
		{
			name:     "single tilde fence backtick-style violation",
			content:  "~~~go\ncode\n~~~\n",
			style:    "backtick",
			wantErrs: []LintError{{File: "test.md", Line: 1, Message: "consistent-code-fence: expected backtick fence, got tilde fence"}},
		},
		{
			name:     "single backtick fence tilde-style violation",
			content:  "```go\ncode\n```\n",
			style:    "tilde",
			wantErrs: []LintError{{File: "test.md", Line: 1, Message: "consistent-code-fence: expected tilde fence, got backtick fence"}},
		},

		// all-backtick documents
		{
			name:     "all backtick consistent no violation",
			content:  "```go\ncode\n```\n\n```python\ncode\n```\n",
			style:    "consistent",
			wantErrs: nil,
		},
		{
			name:     "all backtick backtick-style no violation",
			content:  "```go\ncode\n```\n\n```python\ncode\n```\n",
			style:    "backtick",
			wantErrs: nil,
		},
		{
			name:    "all backtick tilde-style two violations",
			content: "```go\ncode\n```\n\n```python\ncode\n```\n",
			style:   "tilde",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "consistent-code-fence: expected tilde fence, got backtick fence"},
				{File: "test.md", Line: 5, Message: "consistent-code-fence: expected tilde fence, got backtick fence"},
			},
		},

		// all-tilde documents
		{
			name:     "all tilde consistent no violation",
			content:  "~~~go\ncode\n~~~\n\n~~~python\ncode\n~~~\n",
			style:    "consistent",
			wantErrs: nil,
		},
		{
			name:     "all tilde tilde-style no violation",
			content:  "~~~go\ncode\n~~~\n\n~~~python\ncode\n~~~\n",
			style:    "tilde",
			wantErrs: nil,
		},
		{
			name:    "all tilde backtick-style two violations",
			content: "~~~go\ncode\n~~~\n\n~~~python\ncode\n~~~\n",
			style:   "backtick",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "consistent-code-fence: expected backtick fence, got tilde fence"},
				{File: "test.md", Line: 5, Message: "consistent-code-fence: expected backtick fence, got tilde fence"},
			},
		},

		// mixed fences — consistent mode
		{
			name:    "backtick-first consistent: tilde flagged",
			content: "```go\ncode\n```\n\n~~~python\ncode\n~~~\n",
			style:   "consistent",
			wantErrs: []LintError{
				{File: "test.md", Line: 5, Message: "consistent-code-fence: expected backtick fence, got tilde fence"},
			},
		},
		{
			name:    "tilde-first consistent: backtick flagged",
			content: "~~~go\ncode\n~~~\n\n```python\ncode\n```\n",
			style:   "consistent",
			wantErrs: []LintError{
				{File: "test.md", Line: 5, Message: "consistent-code-fence: expected tilde fence, got backtick fence"},
			},
		},
		{
			name:    "consistent multiple violations",
			content: "```go\ncode\n```\n\n~~~python\ncode\n~~~\n\n~~~bash\ncode\n~~~\n",
			style:   "consistent",
			wantErrs: []LintError{
				{File: "test.md", Line: 5, Message: "consistent-code-fence: expected backtick fence, got tilde fence"},
				{File: "test.md", Line: 9, Message: "consistent-code-fence: expected backtick fence, got tilde fence"},
			},
		},

		// content inside code block must not be rechecked
		{
			name:     "tilde-looking content inside backtick block not flagged",
			content:  "```go\n~~~\n```\n",
			style:    "consistent",
			wantErrs: nil,
		},
		{
			name:     "backtick-looking content inside tilde block not flagged",
			content:  "~~~go\n```\n~~~\n",
			style:    "consistent",
			wantErrs: nil,
		},

		// fence inside HTML comment is ignored
		{
			name:     "fence inside HTML comment ignored",
			content:  "text\n<!--\n~~~go\ncode\n~~~\n-->\n```go\ncode\n```\n",
			style:    "consistent",
			wantErrs: nil,
		},
		// fence opener whose info string contains "<!--" must not be skipped
		{
			name:    "fence opener with <!-- in info string not skipped",
			content: "```go <!-- comment -->\ncode\n```\n\n~~~python\ncode\n~~~\n",
			style:   "consistent",
			wantErrs: []LintError{
				{File: "test.md", Line: 5, Message: "consistent-code-fence: expected backtick fence, got tilde fence"},
			},
		},

		// offset
		{
			name:    "offset shifts line numbers",
			content: "```go\ncode\n```\n\n~~~python\ncode\n~~~\n",
			offset:  10,
			style:   "consistent",
			wantErrs: []LintError{
				{File: "test.md", Line: 15, Message: "consistent-code-fence: expected backtick fence, got tilde fence"},
			},
		},

		// invalid style falls back to consistent
		{
			name:    "unknown style falls back to consistent",
			content: "```go\ncode\n```\n\n~~~python\ncode\n~~~\n",
			style:   "unknown",
			wantErrs: []LintError{
				{File: "test.md", Line: 5, Message: "consistent-code-fence: expected backtick fence, got tilde fence"},
			},
		},

		// longer fence markers
		{
			name:     "longer backtick fence consistent no violation",
			content:  "````go\ncode\n````\n\n````python\ncode\n````\n",
			style:    "consistent",
			wantErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(tt.content, "\n")
			got := CheckConsistentCodeFence("test.md", lines, tt.offset, tt.style)

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

package rule

import (
	"strings"
	"testing"
)

func TestCheckSingleH1(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		offset   int
		wantErrs []LintError
	}{
		{
			name:     "single H1 is valid",
			content:  "# Title\n\n## Section\n",
			wantErrs: nil,
		},
		{
			name:     "no H1 is valid",
			content:  "## Section One\n\n## Section Two\n",
			wantErrs: nil,
		},
		{
			name:    "two H1 headings",
			content: "# First\n\nSome content.\n\n# Second\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 5, Message: "Multiple H1 headings found; only one H1 is allowed per file"},
			},
		},
		{
			name:    "three H1 headings",
			content: "# One\n\n# Two\n\n# Three\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "Multiple H1 headings found; only one H1 is allowed per file"},
				{File: "test.md", Line: 5, Message: "Multiple H1 headings found; only one H1 is allowed per file"},
			},
		},
		{
			name:     "H1 inside fenced code block is ignored",
			content:  "# Title\n\n```markdown\n# Not a real heading\n```\n",
			wantErrs: nil,
		},
		{
			name:     "H1 inside tilde fence is ignored",
			content:  "# Title\n\n~~~\n# Ignored\n~~~\n",
			wantErrs: nil,
		},
		{
			name:    "offset shifts line numbers",
			content: "# First\n\n# Second\n",
			offset:  3,
			wantErrs: []LintError{
				{File: "test.md", Line: 6, Message: "Multiple H1 headings found; only one H1 is allowed per file"},
			},
		},
		{
			name:     "empty file",
			content:  "",
			wantErrs: nil,
		},
		{
			name:     "H2 and H3 only",
			content:  "## H2\n\n### H3\n",
			wantErrs: nil,
		},
		{
			name:     "bare hash without space is not H1",
			content:  "# Title\n\n#notaheading\n",
			wantErrs: nil,
		},
		{
			name:    "bare hash alone counts as H1",
			content: "#\n\n#\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "Multiple H1 headings found; only one H1 is allowed per file"},
			},
		},
		{
			name:     "H1 inside block closed by longer fence is ignored",
			content:  "# Title\n\n```markdown\n# Not a heading\n`````\n",
			wantErrs: nil,
		},
		{
			name:    "H1 after block closed by longer fence is detected",
			content: "# Title\n\n```markdown\n# Ignored\n`````\n\n# Second\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 7, Message: "Multiple H1 headings found; only one H1 is allowed per file"},
			},
		},
		{
			name:     "H1 with trailing spaces is still a valid single H1",
			content:  "# Title   \n\n## Section\n",
			wantErrs: nil,
		},
		{
			name:     "H1 with leading spaces is recognized",
			content:  "  # Title\n\n## Section\n",
			wantErrs: nil,
		},
		{
			name:     "backtick-starting non-fence line is ignored",
			content:  "# Title\n\n`inline code` here\n",
			wantErrs: nil,
		},
		{
			name: "fence char inside block that does not close it is ignored",
			// The ``` line inside the ``````-fenced block starts with '`' but is
			// shorter than the opener, so IsClosingFence returns false.
			content:  "# Title\n\n``````go\n```not-closing\n``````\n",
			wantErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(tt.content, "\n")
			got := CheckSingleH1("test.md", lines, tt.offset)

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

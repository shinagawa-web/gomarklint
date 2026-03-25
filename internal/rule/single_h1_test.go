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
			name:     "no headings",
			content:  "Some text.\n",
			wantErrs: nil,
		},
		{
			name:     "single H1",
			content:  "# Title\n\nContent.\n",
			wantErrs: nil,
		},
		{
			name:    "two H1 headings",
			content: "# First\n\n# Second\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "Multiple H1 headings found; only one H1 is allowed per file"},
			},
		},
		{
			name:    "three H1 headings",
			content: "# First\n\n# Second\n\n# Third\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "Multiple H1 headings found; only one H1 is allowed per file"},
				{File: "test.md", Line: 5, Message: "Multiple H1 headings found; only one H1 is allowed per file"},
			},
		},
		{
			name:     "H2 and H3 not counted as H1",
			content:  "## Section\n\n### Subsection\n",
			wantErrs: nil,
		},
		{
			name:     "H1 followed by H2 is valid",
			content:  "# Title\n\n## Section\n",
			wantErrs: nil,
		},
		{
			name:     "H1 inside fenced code block is ignored",
			content:  "# Real Title\n\n```markdown\n# Fake Title\n```\n",
			wantErrs: nil,
		},
		{
			name:     "H1 inside tilde fence is ignored",
			content:  "# Real Title\n\n~~~markdown\n# Fake Title\n~~~\n",
			wantErrs: nil,
		},
		{
			name:    "second H1 after fenced code block is flagged",
			content: "# First\n\n```go\ncode\n```\n\n# Second\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 7, Message: "Multiple H1 headings found; only one H1 is allowed per file"},
			},
		},
		{
			name:     "empty file",
			content:  "",
			wantErrs: nil,
		},
		{
			name:    "offset shifts line numbers",
			content: "# First\n\n# Second\n",
			offset:  5,
			wantErrs: []LintError{
				{File: "test.md", Line: 8, Message: "Multiple H1 headings found; only one H1 is allowed per file"},
			},
		},
		{
			name:     "bare hash is treated as H1",
			content:  "#\n\n# Second\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "Multiple H1 headings found; only one H1 is allowed per file"},
			},
		},
		{
			name:     "hash without space is not H1",
			content:  "#Title\n\n# Real\n",
			wantErrs: nil,
		},
		{
			name:     "zero H1 headings is valid",
			content:  "## Section\n\nContent.\n",
			wantErrs: nil,
		},
		{
			name:     "longer closing fence properly closes block (CommonMark)",
			content:  "# Real Title\n\n```go\ncode\n````\n\n# Second Title\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 7, Message: "Multiple H1 headings found; only one H1 is allowed per file"},
			},
		},
		{
			name:     "H1 inside block not flagged when block closed by longer fence",
			content:  "# Real Title\n\n```markdown\n# Fake Title\n````\n",
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

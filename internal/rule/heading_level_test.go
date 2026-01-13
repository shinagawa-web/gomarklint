package rule

import (
	"testing"
)

func TestCheckHeadingLevels(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		minLevel int
		wantErrs []LintError
	}{
		{
			name:     "valid simple headings",
			content:  "## Heading 2\n### Heading 3\n## Another 2",
			minLevel: 2,
			wantErrs: nil,
		},
		{
			name:     "first heading too low",
			content:  "### Heading 3\n## Heading 2",
			minLevel: 2,
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "First heading should be level 2 (found level 3)"},
			},
		},
		{
			name:     "skip heading level",
			content:  "## Level 2\n#### Level 4 (skipped 3)",
			minLevel: 2,
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "Heading level jumped from 2 to 4"},
			},
		},
		{
			name:     "heading level 1 allowed when minLevel is 1",
			content:  "# Heading 1\n## Heading 2",
			minLevel: 1,
			wantErrs: nil,
		},
		{
			name:     "with frontmatter",
			content:  "---\ntitle: Test\n---\n\n# Heading 1\n## Heading 2",
			minLevel: 2,
			wantErrs: []LintError{
				{File: "test.md", Line: 5, Message: "First heading should be level 2 (found level 1)"},
			},
		},
		{
			name:     "multiple jumps",
			content:  "## Intro\n#### Skip to 4\n### Back to 3\n##### Skip to 5",
			minLevel: 2,
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "Heading level jumped from 2 to 4"},
				{File: "test.md", Line: 4, Message: "Heading level jumped from 3 to 5"},
			},
		},
		{
			name:     "ignore headings in code blocks",
			content:  "## Introduction\n\n```\n# Skipped Level\n\n```\n\n### Normal Again",
			minLevel: 2,
			wantErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckHeadingLevels("test.md", tt.content, tt.minLevel)

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

package rule

import (
	"strings"
	"testing"
)

func TestAtxHeadingLevel(t *testing.T) {
	tests := []struct {
		line string
		want int
	}{
		{"# Heading", 1},
		{"## Heading", 2},
		{"###### Heading", 6},
		{"####### Too deep", 0}, // level > 6
		{"#NoSpace", 0},         // missing space after '#'
		{"#", 1},                // lone '#' with no text (level == len(line))
		{"##\tTab", 2},          // tab after '#'
		{"not a heading", 0},    // no '#'
		{"#\r", 0},              // bare '#' + CRLF remnant: '\r' is not a valid terminator — caller must TrimSpace first
		{"##\r", 0},             // bare '##' + CRLF remnant: same
	}
	for _, tt := range tests {
		got := atxHeadingLevel(tt.line)
		if got != tt.want {
			t.Errorf("atxHeadingLevel(%q) = %d, want %d", tt.line, got, tt.want)
		}
	}
}

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
		{
			name:     "hash without space is not a heading",
			content:  "## Section\n#NotAHeading\n### Subsection",
			minLevel: 2,
			wantErrs: nil,
		},
		{
			name:     "seven hashes is not a heading",
			content:  "## Section\n####### Too deep\n### Subsection",
			minLevel: 2,
			wantErrs: nil,
		},
		{
			name:     "empty lines are skipped",
			content:  "\n\n## Section\n### Subsection",
			minLevel: 2,
			wantErrs: nil,
		},
		{
			name:     "headings inside unclosed code block are ignored",
			content:  "## Section\n```\n# Inside unclosed block\n### Also inside",
			minLevel: 2,
			wantErrs: nil,
		},
		{
			name:     "non-heading lines outside code blocks are skipped",
			content:  "## Section\nSome paragraph text.\n### Subsection",
			minLevel: 2,
			wantErrs: nil,
		},
		{
			name:     "backtick-starting non-fence line is ignored",
			content:  "## Section\n`inline code`\n### Subsection",
			minLevel: 2,
			wantErrs: nil,
		},
		{
			name: "CRLF bare headings are recognized end-to-end",
			// Simulates reading a CRLF file split on "\n": each line retains a
			// trailing '\r'. CheckHeadingLevels calls strings.TrimSpace before
			// atxHeadingLevel, so bare headings like "##\r" are handled correctly.
			content:  "##\r\n###\r",
			minLevel: 2,
			wantErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(tt.content, "\n")
			got := CheckHeadingLevels("test.md", lines, 0, tt.minLevel)

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

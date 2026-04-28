package rule

import (
	"strings"
	"testing"
)

func TestCheckNoTrailingPunctuation(t *testing.T) {
	const defaultPunct = ".,;:!"

	tests := []struct {
		name        string
		content     string
		punctuation string
		offset      int
		wantErrs    []LintError
	}{
		// Default punctuation — each character triggers a violation
		{
			name:        "invalid: trailing period",
			content:     "## Heading.\n",
			punctuation: defaultPunct,
			wantErrs:    []LintError{{File: "test.md", Line: 1, Message: `no-trailing-punctuation: heading ends with "."`}},
		},
		{
			name:        "invalid: trailing comma",
			content:     "## Heading,\n",
			punctuation: defaultPunct,
			wantErrs:    []LintError{{File: "test.md", Line: 1, Message: `no-trailing-punctuation: heading ends with ","`}},
		},
		{
			name:        "invalid: trailing semicolon",
			content:     "## Heading;\n",
			punctuation: defaultPunct,
			wantErrs:    []LintError{{File: "test.md", Line: 1, Message: `no-trailing-punctuation: heading ends with ";"`}},
		},
		{
			name:        "invalid: trailing colon",
			content:     "## Heading:\n",
			punctuation: defaultPunct,
			wantErrs:    []LintError{{File: "test.md", Line: 1, Message: `no-trailing-punctuation: heading ends with ":"`}},
		},
		{
			name:        "invalid: trailing exclamation",
			content:     "## Heading!\n",
			punctuation: defaultPunct,
			wantErrs:    []LintError{{File: "test.md", Line: 1, Message: `no-trailing-punctuation: heading ends with "!"`}},
		},
		// '?' is not in the default set
		{
			name:        "valid: trailing question mark not in default set",
			content:     "## Why does this matter?\n",
			punctuation: defaultPunct,
			wantErrs:    nil,
		},
		// '?' violation when explicitly added to punctuation
		{
			name:        "invalid: trailing question mark when added to set",
			content:     "## Why?\n",
			punctuation: defaultPunct + "?",
			wantErrs:    []LintError{{File: "test.md", Line: 1, Message: `no-trailing-punctuation: heading ends with "?"`}},
		},
		// Valid headings
		{
			name:        "valid: no trailing punctuation",
			content:     "## Clean Heading\n",
			punctuation: defaultPunct,
			wantErrs:    nil,
		},
		{
			name:        "valid: punctuation in the middle",
			content:     "## Config: Options\n",
			punctuation: defaultPunct,
			wantErrs:    nil,
		},
		{
			name:        "valid: ends with closing parenthesis",
			content:     "## Example (Part 1)\n",
			punctuation: defaultPunct,
			wantErrs:    nil,
		},
		{
			name:        "valid: all heading levels",
			content:     "# H1\n## H2\n### H3\n#### H4\n##### H5\n###### H6\n",
			punctuation: defaultPunct,
			wantErrs:    nil,
		},
		// ATX closing sequence
		{
			name:        "invalid: trailing period with closing sequence",
			content:     "## Heading. ##\n",
			punctuation: defaultPunct,
			wantErrs:    []LintError{{File: "test.md", Line: 1, Message: `no-trailing-punctuation: heading ends with "."`}},
		},
		{
			name:        "valid: no trailing punctuation with closing sequence",
			content:     "## Heading ##\n",
			punctuation: defaultPunct,
			wantErrs:    nil,
		},
		// Setext headings
		{
			name:        "invalid: setext H1 with trailing period",
			content:     "Heading.\n========\n",
			punctuation: defaultPunct,
			wantErrs:    []LintError{{File: "test.md", Line: 1, Message: `no-trailing-punctuation: heading ends with "."`}},
		},
		{
			name:        "invalid: setext H2 with trailing colon",
			content:     "Section:\n--------\n",
			punctuation: defaultPunct,
			wantErrs:    []LintError{{File: "test.md", Line: 1, Message: `no-trailing-punctuation: heading ends with ":"`}},
		},
		{
			name:        "valid: setext heading without trailing punctuation",
			content:     "Heading\n=======\n",
			punctuation: defaultPunct,
			wantErrs:    nil,
		},
		// Code block — headings inside must be ignored
		{
			name:        "valid: heading with punctuation inside fenced code block (backtick)",
			content:     "```\n## Inside.\n```\n",
			punctuation: defaultPunct,
			wantErrs:    nil,
		},
		{
			name:        "valid: heading with punctuation inside fenced code block (tilde)",
			content:     "~~~\n## Inside!\n~~~\n",
			punctuation: defaultPunct,
			wantErrs:    nil,
		},
		{
			name:        "valid: setext text line inside fenced code block",
			content:     "```\nHeading.\n========\n```\n",
			punctuation: defaultPunct,
			wantErrs:    nil,
		},
		// Heading immediately after code block close
		{
			name:        "invalid: heading after code block close",
			content:     "```\ncode\n```\n## After.\n",
			punctuation: defaultPunct,
			wantErrs:    []LintError{{File: "test.md", Line: 4, Message: `no-trailing-punctuation: heading ends with "."`}},
		},
		// Offset handling
		{
			name:        "offset shifts line number",
			content:     "## Heading.\n",
			punctuation: defaultPunct,
			offset:      3,
			wantErrs:    []LintError{{File: "test.md", Line: 4, Message: `no-trailing-punctuation: heading ends with "."`}},
		},
		// Multiple violations
		{
			name:        "multiple violations",
			content:     "## First.\n\n## Second!\n",
			punctuation: defaultPunct,
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: `no-trailing-punctuation: heading ends with "."`},
				{File: "test.md", Line: 3, Message: `no-trailing-punctuation: heading ends with "!"`},
			},
		},
		// Empty punctuation set — no violations
		{
			name:        "empty punctuation set disables rule",
			content:     "## Heading.\n",
			punctuation: "",
			wantErrs:    nil,
		},
		// Empty file
		{
			name:        "empty file",
			content:     "",
			punctuation: defaultPunct,
			wantErrs:    nil,
		},
		// Heading with no text
		{
			name:        "valid: empty heading",
			content:     "##\n",
			punctuation: defaultPunct,
			wantErrs:    nil,
		},
		// Setext underline after blank line is a thematic break, not a heading
		{
			name:        "valid: dashes after blank line are not a setext heading",
			content:     "Paragraph\n\n---\n",
			punctuation: defaultPunct,
			wantErrs:    nil,
		},
		// Setext underline after list item is not a heading
		{
			name:        "valid: dashes after list item are not a setext heading",
			content:     "- item.\n---\n",
			punctuation: defaultPunct,
			wantErrs:    nil,
		},
		// Line starting with '-' after a paragraph but not a setext underline
		{
			name:        "valid: list item after paragraph is not a setext underline",
			content:     "Paragraph text\n- list item\n",
			punctuation: defaultPunct,
			wantErrs:    nil,
		},
		// ATX heading whose text is only the closing '#' sequence (j < 0 branch)
		{
			name:        "valid: heading whose content is only closing hashes",
			content:     "## ##\n",
			punctuation: defaultPunct,
			wantErrs:    nil,
		},
		// Line starting with '#' but no space after — not a valid ATX heading
		{
			name:        "valid: hash without space is not a heading",
			content:     "##nospace\n",
			punctuation: defaultPunct,
			wantErrs:    nil,
		},
		// Blockquote after paragraph — not a setext underline
		{
			name:        "valid: blockquote after paragraph is not a setext underline",
			content:     "Paragraph text\n> quoted\n",
			punctuation: defaultPunct,
			wantErrs:    nil,
		},
		// Ordered list after paragraph — not a setext underline
		{
			name:        "valid: ordered list after paragraph is not a setext underline",
			content:     "Paragraph text\n1. item\n",
			punctuation: defaultPunct,
			wantErrs:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(tt.content, "\n")
			got := CheckNoTrailingPunctuation("test.md", lines, tt.offset, tt.punctuation)

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

package rule

import (
	"testing"
)

func TestCheckNoSetextHeadings(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantErrs []LintError
	}{
		{
			name:     "horizontal rule",
			content:  "Paragraph text\n\n---\n\nNext paragraph text",
			wantErrs: nil,
		},
		{
			name:     "horizontal rule after whitespace-only line",
			content:  "paragraph text\n    \n-----",
			wantErrs: nil,
		},
		{
			name:     "horizontal rule at start of file",
			content:  "-----\n",
			wantErrs: nil,
		},
		{
			name:     "ignore code blocks",
			content:  "```\nI am in a code block\n-----\n```",
			wantErrs: nil,
		},
		{
			name:    "forbid setext headings with dashes",
			content: "I am an h2 heading\n---",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "Setext heading found (prefer ATX style instead)"},
			},
		},
		{
			name:    "forbid setext headings with dashes and trailing whitespace",
			content: "I am an h2 heading\n---  ",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "Setext heading found (prefer ATX style instead)"},
			},
		},
		{
			name:    "forbid setext headings with equals",
			content: "I am an h1 heading\n===",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "Setext heading found (prefer ATX style instead)"},
			},
		},
		{
			name:    "handle spaces before heading text",
			content: "   I am an h1 heading\n===",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "Setext heading found (prefer ATX style instead)"},
			},
		},
		{
			name:     "ignore mixed underline characters",
			content:  "I am not a heading\n=-=-=-",
			wantErrs: nil,
		},
		{
			name:     "ignore non-underline characters",
			content:  "I am not a heading\nAnd neither am I",
			wantErrs: nil,
		},
		{
			name:     "list item followed by hr should not be a heading",
			content:  "- list item\n---",
			wantErrs: nil,
		},
		{
			name:    "multiple spaces before underline",
			content: "Heading\n   ===",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "Setext heading found (prefer ATX style instead)"},
			},
		},
		{
			name:     "too many spaces before underline is a code block",
			content:  "Not a heading\n    ===",
			wantErrs: nil,
		},
		{
			name:    "setext heading with multiple lines of text",
			content: "This is a\nmulti-line heading\n===",
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "Setext heading found (prefer ATX style instead)"},
			},
		},
		{
			name:     "blockquote followed by hr should not be a heading",
			content:  "> This is a blockquote\n---",
			wantErrs: nil,
		},
		{
			name:     "nested blockquote followed by hr should not be a heading",
			content:  ">> Nested quote\n===",
			wantErrs: nil,
		},
		{
			name:     "lazy blockquote followed by hr should not be a heading",
			content:  "> This is a\nquote\n---",
			wantErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckNoSetextHeadings("test.md", tt.content)

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

package rule

import (
	"testing"
)

func TestCheckNoMultipleBlankLines(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantErrs []LintError
	}{
		{
			name:     "no consecutive blank lines",
			content:  "# Heading\n\nParagraph\n\nAnother paragraph\n",
			wantErrs: nil,
		},
		{
			name:    "two consecutive blank lines",
			content: "# Heading\n\n\nParagraph\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "Multiple consecutive blank lines"},
			},
		},
		{
			name:    "three consecutive blank lines",
			content: "# Heading\n\n\n\nParagraph\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "Multiple consecutive blank lines"},
				{File: "test.md", Line: 4, Message: "Multiple consecutive blank lines"},
			},
		},
		{
			name:    "multiple occurrences",
			content: "# Heading\n\n\nParagraph\n\n\nAnother\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "Multiple consecutive blank lines"},
				{File: "test.md", Line: 6, Message: "Multiple consecutive blank lines"},
			},
		},
		{
			name:    "with frontmatter",
			content: "---\ntitle: Test\n---\n\n# Heading\n\n\nParagraph\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 7, Message: "Multiple consecutive blank lines"},
			},
		},
		{
			name:     "single line",
			content:  "# Heading\n",
			wantErrs: nil,
		},
		{
			name:    "blank lines with spaces",
			content: "# Heading\n  \n  \nParagraph\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "Multiple consecutive blank lines"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckNoMultipleBlankLines("test.md", tt.content)

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

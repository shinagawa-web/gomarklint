package rule

import (
	"testing"
)

func TestCheckFinalBlankLine(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantErrs []LintError
	}{
		{
			name:     "with blank line",
			content:  "# Hello\nWorld\n",
			wantErrs: nil,
		},
		{
			name:    "no blank line",
			content: "# Hello\nWorld",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "Missing final blank line"},
			},
		},
		{
			name:    "empty file",
			content: "",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "Missing final blank line"},
			},
		},
		{
			name:     "single newline",
			content:  "\n",
			wantErrs: nil,
		},
		{
			name:     "frontmatter with blank",
			content:  "---\ntitle: Test\n---\n\n# Hello\n\n",
			wantErrs: nil,
		},
		{
			name:    "frontmatter no blank",
			content: "---\ntitle: Test\n---\n\n# Hello",
			wantErrs: []LintError{
				{File: "test.md", Line: 5, Message: "Missing final blank line"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckFinalBlankLine("test.md", tt.content)

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

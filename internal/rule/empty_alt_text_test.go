package rule

import (
	"testing"
)

func TestCheckEmptyAltText(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantErrs []LintError
	}{
		{
			name:     "no images",
			content:  "# Hello\nThis is a test.",
			wantErrs: nil,
		},
		{
			name:     "image with alt text",
			content:  "![logo](https://example.com/logo.png)",
			wantErrs: nil,
		},
		{
			name:    "image with empty alt text",
			content: "![](https://example.com/image.png)",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "image with empty alt text"},
			},
		},
		{
			name:    "mixed lines with one empty alt",
			content: "\n![a](url)\n![](url)\n![b](url)",
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "image with empty alt text"},
			},
		},
		{
			name:    "multiple empty alt text",
			content: "![](img1.png)\nSome text\n![](img2.png)\n![desc](img3.png)\n![](img4.png)",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "image with empty alt text"},
				{File: "test.md", Line: 3, Message: "image with empty alt text"},
				{File: "test.md", Line: 5, Message: "image with empty alt text"},
			},
		},
		{
			name:    "image with spaces in alt text",
			content: "![  ](image.png)",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "image with empty alt text"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckEmptyAltText("test.md", tt.content)

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

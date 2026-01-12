package rule

import (
	"os"
	"testing"
)

func TestCheckNoMultipleBlankLines(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		wantErrors int
	}{
		{
			name:       "no consecutive blank lines",
			content:    "# Heading\n\nParagraph\n\nAnother paragraph\n",
			wantErrors: 0,
		},
		{
			name:       "two consecutive blank lines",
			content:    "# Heading\n\n\nParagraph\n",
			wantErrors: 1,
		},
		{
			name:       "three consecutive blank lines",
			content:    "# Heading\n\n\n\nParagraph\n",
			wantErrors: 2,
		},
		{
			name:       "multiple occurrences",
			content:    "# Heading\n\n\nParagraph\n\n\nAnother\n",
			wantErrors: 2,
		},
		{
			name:       "with frontmatter",
			content:    "---\ntitle: Test\n---\n\n# Heading\n\n\nParagraph\n",
			wantErrors: 1,
		},
		{
			name:       "single line",
			content:    "# Heading\n",
			wantErrors: 0,
		},
		{
			name:       "blank lines with spaces",
			content:    "# Heading\n  \n  \nParagraph\n",
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := CheckNoMultipleBlankLines("test.md", tt.content)
			if len(errors) != tt.wantErrors {
				t.Errorf("expected %d errors, got %d: %v", tt.wantErrors, len(errors), errors)
			}
		})
	}
}

func TestCheckNoMultipleBlankLinesWithFiles(t *testing.T) {
	tests := []struct {
		name       string
		filepath   string
		wantErrors int
	}{
		{"valid markdown", "testdata/sample.md", 0},
		{"with links", "testdata/sample_links.md", 1}, // has one occurrence of multiple blank lines
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := getTestFilePath(tt.filepath)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Skipf("test file not found: %v", err)
			}
			errors := CheckNoMultipleBlankLines(path, string(data))
			if len(errors) != tt.wantErrors {
				t.Errorf("expected %d errors, got %d: %v", tt.wantErrors, len(errors), errors)
			}
		})
	}
}

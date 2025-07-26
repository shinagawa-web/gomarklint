package rule

import (
	"reflect"
	"testing"
)

func TestCheckEmptyAltText(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []LintError
	}{
		{
			name:     "no images",
			content:  "# Hello\nThis is a test.",
			expected: nil,
		},
		{
			name:     "image with alt text",
			content:  "![logo](https://example.com/logo.png)",
			expected: nil,
		},
		{
			name:    "image with empty alt text",
			content: "![](https://example.com/image.png)",
			expected: []LintError{
				{File: "test.md", Line: 1, Message: "image with empty alt text"},
			},
		},
		{
			name: "mixed lines",
			content: `
![a](url)
![](url)
![b](url)`,
			expected: []LintError{
				{File: "test.md", Line: 3, Message: "image with empty alt text"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckEmptyAltText("test.md", tt.content)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("CheckEmptyAltText() = %v, want %v", got, tt.expected)
			}
		})
	}
}

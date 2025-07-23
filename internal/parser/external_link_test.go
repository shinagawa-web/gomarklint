package parser

import (
	"reflect"
	"sort"
	"testing"
)

func TestExtractExternalLinks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "inline link",
			input: `
This is a [link](https://example.com/page).
`,
			expected: []string{"https://example.com/page"},
		},
		{
			name: "image link",
			input: `
Here is an image: ![logo](https://example.com/logo.png)
`,
			expected: []string{"https://example.com/logo.png"},
		},
		{
			name: "bare url",
			input: `
Check this out: https://example.com/docs
`,
			expected: []string{"https://example.com/docs"},
		},
		{
			name: "duplicates",
			input: `
[Link1](https://dup.com)
[Link2](https://dup.com)
https://dup.com
`,
			expected: []string{"https://dup.com"},
		},
		{
			name: "non-http links ignored",
			input: `
[Local](/relative/path)
[FTP](ftp://example.com)
`,
			expected: []string{},
		},
		{
			name: "mixed valid and invalid",
			input: `
Here's something: [Valid](https://valid.com)
Here's something: [Local](/not-checked)
https://also.valid.com
`,
			expected: []string{"https://valid.com", "https://also.valid.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractExternalLinks(tt.input)

			// 並び順を保証しないためソートして比較
			sort.Strings(got)
			sort.Strings(tt.expected)

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("got = %v, want = %v", got, tt.expected)
			}
		})
	}
}

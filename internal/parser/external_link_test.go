package parser_test

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/shinagawa-web/gomarklint/internal/parser"
)

func TestExtractExternalLinksWithLineNumbers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []parser.ExtractedLink
	}{
		{
			name: "inline link",
			input: `
This is a [link](https://example.com/page).
`,
			expected: []parser.ExtractedLink{
				{URL: "https://example.com/page", Line: 2},
			},
		},
		{
			name: "image link",
			input: `
Here is an image: ![logo](https://example.com/logo.png)
`,
			expected: []parser.ExtractedLink{
				{URL: "https://example.com/logo.png", Line: 2},
			},
		},
		{
			name: "bare url",
			input: `
Check this out: https://example.com/docs
`,
			expected: []parser.ExtractedLink{
				{URL: "https://example.com/docs", Line: 2},
			},
		},
		{
			name: "duplicates (returns all occurrences)",
			input: `
[Link1](https://dup.com)
[Link2](https://dup.com)
https://dup.com
`,
			expected: []parser.ExtractedLink{
				{URL: "https://dup.com", Line: 2},
				{URL: "https://dup.com", Line: 3},
				{URL: "https://dup.com", Line: 4},
			},
		},
		{
			name: "non-http links ignored",
			input: `
[Local](/relative/path)
[FTP](ftp://example.com)
`,
			expected: nil,
		},
		{
			name: "mixed valid and invalid",
			input: `
Here's something: [Valid](https://valid.com)
Here's something: [Local](/not-checked)
https://also.valid.com
`,
			expected: []parser.ExtractedLink{
				{URL: "https://valid.com", Line: 2},
				{URL: "https://also.valid.com", Line: 4},
			},
		},
		{
			name: "same URL in same line (inline and bare) - deduplicated",
			input: `
Check [this](https://example.com) and also https://example.com
`,
			expected: []parser.ExtractedLink{
				{URL: "https://example.com", Line: 2},
			},
		},
		{
			name: "multiple different URLs in same line",
			input: `
See [link1](https://first.com) and [link2](https://second.com)
`,
			expected: []parser.ExtractedLink{
				{URL: "https://first.com", Line: 2},
				{URL: "https://second.com", Line: 2},
			},
		},
		{
			name: "http and https",
			input: `
[Secure](https://secure.example.com)
[Insecure](http://insecure.example.com)
`,
			expected: []parser.ExtractedLink{
				{URL: "https://secure.example.com", Line: 2},
				{URL: "http://insecure.example.com", Line: 3},
			},
		},
		{
			name:     "empty input",
			input:    ``,
			expected: nil,
		},
		{
			name: "only whitespace",
			input: `

   

`,
			expected: nil,
		},
		{
			name: "image and inline link with same URL in same line",
			input: `
![image](https://example.com/img.png) and [link](https://example.com/img.png)
`,
			expected: []parser.ExtractedLink{
				{URL: "https://example.com/img.png", Line: 2},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(tt.input, "\n")
			offset := 0
			got := parser.ExtractExternalLinksWithLineNumbers(lines, offset)

			// Sort to ensure order-independent comparison
			sort.Slice(got, func(i, j int) bool {
				return got[i].URL < got[j].URL
			})
			sort.Slice(tt.expected, func(i, j int) bool {
				return tt.expected[i].URL < tt.expected[j].URL
			})

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("got = %#v, want = %#v", got, tt.expected)
			}
		})
	}
}

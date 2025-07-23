package parser_test

import (
	"reflect"
	"sort"
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
			name: "duplicates (should only take first occurrence)",
			input: `
[Link1](https://dup.com)
[Link2](https://dup.com)
https://dup.com
`,
			expected: []parser.ExtractedLink{
				{URL: "https://dup.com", Line: 2},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.ExtractExternalLinksWithLineNumbers(tt.input)

			// 並び順が関係しないようにソート（複数URLのケースに備えて）
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

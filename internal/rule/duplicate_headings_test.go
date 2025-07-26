package rule

import (
	"reflect"
	"testing"
)

func TestCheckDuplicateHeadings(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []LintError
	}{
		{
			name: "no duplicates",
			content: `# Heading 1
Some text
## Heading 2
### Heading 3`,
			expected: nil,
		},
		{
			name: "duplicates at same level",
			content: `## Introduction
Text here
## Introduction`,
			expected: []LintError{
				{File: "test.md", Line: 3, Message: `duplicate heading: "introduction"`},
			},
		},
		{
			name: "duplicates across different levels",
			content: `# Overview
## Overview`,
			expected: []LintError{
				{File: "test.md", Line: 2, Message: `duplicate heading: "overview"`},
			},
		},
		{
			name: "case insensitive duplicates",
			content: `## Summary
## summary`,
			expected: []LintError{
				{File: "test.md", Line: 2, Message: `duplicate heading: "summary"`},
			},
		},
		{
			name:    "duplicates with trailing spaces",
			content: "## Details\n## Details \n## Details　", // Note: this includes a full-width space
			expected: []LintError{
				{File: "test.md", Line: 2, Message: `duplicate heading: "details"`},
				{File: "test.md", Line: 3, Message: `duplicate heading: "details"`},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckDuplicateHeadings("test.md", tt.content)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("CheckDuplicateHeadings() = %v, want %v", got, tt.expected)
			}
		})
	}
}

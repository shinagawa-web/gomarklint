package rule

import (
	"testing"
)

func TestCheckDuplicateHeadings(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantErrs []LintError
	}{
		{
			name:     "no duplicates",
			content:  "# Heading 1\nSome text\n## Heading 2\n### Heading 3",
			wantErrs: nil,
		},
		{
			name:    "duplicates at same level",
			content: "## Introduction\nText here\n## Introduction",
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: `duplicate heading: "introduction"`},
			},
		},
		{
			name:    "duplicates across different levels",
			content: "# Overview\n## Overview",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: `duplicate heading: "overview"`},
			},
		},
		{
			name:    "case insensitive duplicates",
			content: "## Summary\n## summary",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: `duplicate heading: "summary"`},
			},
		},
		{
			name:    "duplicates with trailing spaces",
			content: "## Details\n## Details \n## Details　", // Note: includes full-width space
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: `duplicate heading: "details"`},
				{File: "test.md", Line: 3, Message: `duplicate heading: "details"`},
			},
		},
		{
			name:    "multiple duplicates",
			content: "# Intro\n## Intro\n## Content\n## Content\n## Intro",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: `duplicate heading: "intro"`},
				{File: "test.md", Line: 4, Message: `duplicate heading: "content"`},
				{File: "test.md", Line: 5, Message: `duplicate heading: "intro"`},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckDuplicateHeadings("test.md", tt.content)

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

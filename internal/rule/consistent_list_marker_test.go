package rule

import (
	"strings"
	"testing"
)

func TestCheckConsistentListMarker(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		offset   int
		style    string
		wantErrs []LintError
	}{
		// empty / no lists
		{
			name:     "empty doc",
			content:  "",
			style:    "consistent",
			wantErrs: nil,
		},
		{
			name:     "no lists",
			content:  "# Hello\n\nSome paragraph.\n",
			style:    "consistent",
			wantErrs: nil,
		},

		// single item — never flagged in consistent mode
		{
			name:     "single dash consistent",
			content:  "- item\n",
			style:    "consistent",
			wantErrs: nil,
		},
		{
			name:     "single asterisk consistent",
			content:  "* item\n",
			style:    "consistent",
			wantErrs: nil,
		},
		{
			name:     "single plus consistent",
			content:  "+ item\n",
			style:    "consistent",
			wantErrs: nil,
		},

		// single item — style violations
		{
			name:    "single asterisk dash-style violation",
			content: "* item\n",
			style:   "dash",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "consistent-list-marker: expected dash marker, got asterisk marker"},
			},
		},
		{
			name:    "single plus dash-style violation",
			content: "+ item\n",
			style:   "dash",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "consistent-list-marker: expected dash marker, got plus marker"},
			},
		},
		{
			name:    "single dash asterisk-style violation",
			content: "- item\n",
			style:   "asterisk",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "consistent-list-marker: expected asterisk marker, got dash marker"},
			},
		},
		{
			name:    "single dash plus-style violation",
			content: "- item\n",
			style:   "plus",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "consistent-list-marker: expected plus marker, got dash marker"},
			},
		},

		// all-dash documents
		{
			name:     "all dash consistent no violation",
			content:  "- one\n- two\n- three\n",
			style:    "consistent",
			wantErrs: nil,
		},
		{
			name:     "all dash dash-style no violation",
			content:  "- one\n- two\n- three\n",
			style:    "dash",
			wantErrs: nil,
		},
		{
			name:    "all dash asterisk-style violations",
			content: "- one\n- two\n",
			style:   "asterisk",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "consistent-list-marker: expected asterisk marker, got dash marker"},
				{File: "test.md", Line: 2, Message: "consistent-list-marker: expected asterisk marker, got dash marker"},
			},
		},

		// mixed markers — consistent mode
		{
			name:    "dash-first consistent: asterisk flagged",
			content: "- one\n* two\n",
			style:   "consistent",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "consistent-list-marker: expected dash marker, got asterisk marker"},
			},
		},
		{
			name:    "asterisk-first consistent: dash flagged",
			content: "* one\n- two\n",
			style:   "consistent",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "consistent-list-marker: expected asterisk marker, got dash marker"},
			},
		},
		{
			name:    "plus-first consistent: others flagged",
			content: "+ one\n- two\n* three\n",
			style:   "consistent",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "consistent-list-marker: expected plus marker, got dash marker"},
				{File: "test.md", Line: 3, Message: "consistent-list-marker: expected plus marker, got asterisk marker"},
			},
		},

		// ordered list items must not be flagged
		{
			name:     "ordered list not flagged",
			content:  "1. first\n2. second\n",
			style:    "dash",
			wantErrs: nil,
		},

		// lines that look like list markers but aren't
		{
			name:     "asterisk not followed by space is not a list item",
			content:  "*not a list*\n",
			style:    "dash",
			wantErrs: nil,
		},
		{
			name:     "dash at end of line is not a list item",
			content:  "-\n",
			style:    "asterisk",
			wantErrs: nil,
		},
		{
			name:     "marker followed by only spaces is not a list item",
			content:  "-   \n",
			style:    "asterisk",
			wantErrs: nil,
		},
		{
			name:     "marker followed by tab then nothing is not a list item",
			content:  "-\t\n",
			style:    "asterisk",
			wantErrs: nil,
		},
		{
			name:     "CRLF line with only CR after marker+space is not a list item",
			content:  "- \r\n",
			style:    "asterisk",
			wantErrs: nil,
		},

		// valid list items with multiple spaces or tab after marker
		{
			name:    "marker followed by multiple spaces then text is a list item",
			content: "-   item\n",
			style:   "asterisk",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "consistent-list-marker: expected asterisk marker, got dash marker"},
			},
		},
		{
			name:    "marker followed by tab then text is a list item",
			content: "-\titem\n",
			style:   "asterisk",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "consistent-list-marker: expected asterisk marker, got dash marker"},
			},
		},

		// indented list items
		{
			name:     "indented dash consistent no violation",
			content:  "- one\n  - nested\n",
			style:    "consistent",
			wantErrs: nil,
		},
		{
			name:    "indented mixed consistent: nested flagged",
			content: "- one\n  * nested\n",
			style:   "consistent",
			wantErrs: []LintError{
				{File: "test.md", Line: 2, Message: "consistent-list-marker: expected dash marker, got asterisk marker"},
			},
		},

		// fenced code block content must be skipped
		{
			name:     "list marker inside code block not flagged",
			content:  "- item\n\n```\n* not a list\n```\n",
			style:    "consistent",
			wantErrs: nil,
		},

		// offset
		{
			name:    "offset shifts line numbers",
			content: "- one\n* two\n",
			offset:  10,
			style:   "consistent",
			wantErrs: []LintError{
				{File: "test.md", Line: 12, Message: "consistent-list-marker: expected dash marker, got asterisk marker"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(tt.content, "\n")
			got := CheckConsistentListMarker("test.md", lines, tt.offset, tt.style)

			if len(got) != len(tt.wantErrs) {
				t.Fatalf("got %d errors, want %d\ngot:  %v\nwant: %v", len(got), len(tt.wantErrs), got, tt.wantErrs)
			}
			for i := range got {
				if got[i] != tt.wantErrs[i] {
					t.Errorf("error[%d]: got %+v, want %+v", i, got[i], tt.wantErrs[i])
				}
			}
		})
	}
}

func TestListItemMarker(t *testing.T) {
	// all-space line: exhausts the leading-whitespace loop, hits i >= len(line) guard
	if ch, ok := listItemMarker("   "); ok {
		t.Errorf("expected false for all-space line, got ch=%q ok=%v", ch, ok)
	}
}

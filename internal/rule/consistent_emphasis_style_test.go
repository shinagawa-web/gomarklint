package rule

import (
	"strings"
	"testing"
)

func TestCheckConsistentEmphasisStyle(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		offset   int
		style    string
		wantErrs []LintError
	}{
		// empty / no emphasis
		{
			name:     "empty doc",
			content:  "",
			style:    "consistent",
			wantErrs: nil,
		},
		{
			name:     "no emphasis",
			content:  "# Hello\n\nSome paragraph.\n",
			style:    "consistent",
			wantErrs: nil,
		},

		// single emphasis — never flagged in consistent mode
		{
			name:     "single asterisk consistent",
			content:  "This is *italic* text.\n",
			style:    "consistent",
			wantErrs: nil,
		},
		{
			name:     "single underscore consistent",
			content:  "This is _italic_ text.\n",
			style:    "consistent",
			wantErrs: nil,
		},
		{
			name:    "single underscore asterisk-style violation",
			content: "This is _italic_ text.\n",
			style:   "asterisk",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "consistent-emphasis-style: expected asterisk emphasis, got underscore emphasis"},
			},
		},
		{
			name:    "single asterisk underscore-style violation",
			content: "This is *italic* text.\n",
			style:   "underscore",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "consistent-emphasis-style: expected underscore emphasis, got asterisk emphasis"},
			},
		},

		// strong emphasis
		{
			name:     "strong asterisk consistent no violation",
			content:  "This is **bold** text.\n",
			style:    "consistent",
			wantErrs: nil,
		},
		{
			name:    "strong underscore asterisk-style violation",
			content: "This is __bold__ text.\n",
			style:   "asterisk",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "consistent-emphasis-style: expected asterisk emphasis, got underscore emphasis"},
			},
		},

		// all-asterisk documents
		{
			name:     "all asterisk consistent no violation",
			content:  "This is *italic* text.\n\nThis is **bold** text.\n",
			style:    "consistent",
			wantErrs: nil,
		},
		{
			name:     "all asterisk asterisk-style no violation",
			content:  "This is *italic* text.\n\nThis is **bold** text.\n",
			style:    "asterisk",
			wantErrs: nil,
		},
		{
			name:    "all asterisk underscore-style two violations",
			content: "This is *italic* text.\n\nThis is **bold** text.\n",
			style:   "underscore",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "consistent-emphasis-style: expected underscore emphasis, got asterisk emphasis"},
				{File: "test.md", Line: 3, Message: "consistent-emphasis-style: expected underscore emphasis, got asterisk emphasis"},
			},
		},

		// all-underscore documents
		{
			name:     "all underscore consistent no violation",
			content:  "This is _italic_ text.\n\nThis is __bold__ text.\n",
			style:    "consistent",
			wantErrs: nil,
		},
		{
			name:     "all underscore underscore-style no violation",
			content:  "This is _italic_ text.\n\nThis is __bold__ text.\n",
			style:    "underscore",
			wantErrs: nil,
		},
		{
			name:    "all underscore asterisk-style two violations",
			content: "This is _italic_ text.\n\nThis is __bold__ text.\n",
			style:   "asterisk",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "consistent-emphasis-style: expected asterisk emphasis, got underscore emphasis"},
				{File: "test.md", Line: 3, Message: "consistent-emphasis-style: expected asterisk emphasis, got underscore emphasis"},
			},
		},

		// mixed — consistent mode
		{
			name:    "asterisk-first consistent: underscore flagged",
			content: "This is *italic* text.\n\nThis is _also italic_ text.\n",
			style:   "consistent",
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "consistent-emphasis-style: expected asterisk emphasis, got underscore emphasis"},
			},
		},
		{
			name:    "underscore-first consistent: asterisk flagged",
			content: "This is _italic_ text.\n\nThis is *also italic* text.\n",
			style:   "consistent",
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "consistent-emphasis-style: expected underscore emphasis, got asterisk emphasis"},
			},
		},
		{
			name:    "consistent multiple violations",
			content: "This is *italic* text.\n\nThis is _also italic_ text.\n\nThis is __bold__ text.\n",
			style:   "consistent",
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "consistent-emphasis-style: expected asterisk emphasis, got underscore emphasis"},
				{File: "test.md", Line: 5, Message: "consistent-emphasis-style: expected asterisk emphasis, got underscore emphasis"},
			},
		},

		// mid-word underscores must not be flagged
		{
			name:     "snake_case not treated as emphasis",
			content:  "Use snake_case naming.\n",
			style:    "asterisk",
			wantErrs: nil,
		},
		{
			name:     "multiple underscores in identifier not flagged",
			content:  "The foo_bar_baz variable.\n",
			style:    "asterisk",
			wantErrs: nil,
		},
		{
			name:    "leading underscore emphasis not mid-word",
			content: "_italic_ and snake_case.\n",
			style:   "asterisk",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "consistent-emphasis-style: expected asterisk emphasis, got underscore emphasis"},
			},
		},

		// content inside code block must not be rechecked
		{
			name:     "emphasis inside fenced code block not flagged",
			content:  "```\n_italic_\n```\n",
			style:    "asterisk",
			wantErrs: nil,
		},
		{
			name:     "emphasis inside inline code not flagged",
			content:  "Use `_italic_` syntax.\n",
			style:    "asterisk",
			wantErrs: nil,
		},

		// delimiter followed by whitespace is not an opener
		{
			name:     "asterisk followed by space is not an opener",
			content:  "Use * as a bullet.\n",
			style:    "underscore",
			wantErrs: nil,
		},

		// escaped markers
		{
			name:     "escaped asterisk not treated as emphasis",
			content:  "Use \\*escaped\\* asterisk.\n",
			style:    "underscore",
			wantErrs: nil,
		},
		{
			name:     "escaped underscore not treated as emphasis",
			content:  "Use \\_escaped\\_ underscore.\n",
			style:    "asterisk",
			wantErrs: nil,
		},

		// triple+ runs not treated as emphasis
		{
			name:     "triple asterisk run not treated as emphasis opener",
			content:  "***not emphasis***\n",
			style:    "underscore",
			wantErrs: nil,
		},

		// closing delimiter followed by punctuation must not be double-counted
		{
			name:    "closing underscore before period not counted as second opener",
			content: "This is _italic_. More text.\n",
			style:   "asterisk",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "consistent-emphasis-style: expected asterisk emphasis, got underscore emphasis"},
			},
		},
		{
			name:    "closing asterisk before comma not counted as second opener",
			content: "Use *bold*, not plain.\n",
			style:   "underscore",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "consistent-emphasis-style: expected underscore emphasis, got asterisk emphasis"},
			},
		},
		{
			name:     "unclosed emphasis not counted",
			content:  "This has *no closer.\n",
			style:    "underscore",
			wantErrs: nil,
		},
		{
			name:    "wrong-length run skipped, correct closer found",
			content: "Use *italic**more* text.\n",
			style:   "underscore",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "consistent-emphasis-style: expected underscore emphasis, got asterisk emphasis"},
			},
		},

		// offset
		{
			name:    "offset shifts line numbers",
			content: "This is *italic* text.\n\nThis is _also italic_ text.\n",
			offset:  10,
			style:   "consistent",
			wantErrs: []LintError{
				{File: "test.md", Line: 13, Message: "consistent-emphasis-style: expected asterisk emphasis, got underscore emphasis"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(tt.content, "\n")
			got := CheckConsistentEmphasisStyle("test.md", lines, tt.offset, tt.style)

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

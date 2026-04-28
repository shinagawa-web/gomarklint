package rule

import (
	"strings"
	"testing"
)

func TestCheckNoEmphasisAsHeading(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		offset   int
		wantErrs []LintError
	}{
		// ── valid cases ──────────────────────────────────────────────────────
		{
			name:     "valid: ATX heading",
			content:  "## Section\n",
			wantErrs: nil,
		},
		{
			name:     "valid: plain text",
			content:  "This is some text.\n",
			wantErrs: nil,
		},
		{
			name:     "valid: bold ending with period",
			content:  "**This is a sentence.**\n",
			wantErrs: nil,
		},
		{
			name:     "valid: italic ending with comma",
			content:  "*Some clause,*\n",
			wantErrs: nil,
		},
		{
			name:     "valid: bold ending with question mark",
			content:  "**Are you sure?**\n",
			wantErrs: nil,
		},
		{
			name:     "valid: bold ending with exclamation",
			content:  "**Watch out!**\n",
			wantErrs: nil,
		},
		{
			name:     "valid: emphasis inside fenced code block",
			content:  "```\n**Heading**\n```\n",
			wantErrs: nil,
		},
		{
			name:     "valid: emphasis inside tilde fence",
			content:  "~~~\n**Heading**\n~~~\n",
			wantErrs: nil,
		},
		{
			name:     "valid: emphasis inside inline code",
			content:  "Use `**bold**` for styling.\n",
			wantErrs: nil,
		},
		{
			name:     "invalid: backtick-leading non-fence line does not bypass later violation",
			content:  "`inline code`\n**Heading**\n",
			wantErrs: []LintError{{File: "test.md", Line: 2, Message: "no-emphasis-as-heading: emphasis used as heading, use ATX heading instead: **Heading**"}},
		},
		{
			name:     "valid: delimiter-only line too short to be an emphasis span",
			content:  "**\n",
			wantErrs: nil,
		},
		{
			name:     "valid: bold with surrounding text",
			content:  "This is **important** text.\n",
			wantErrs: nil,
		},
		{
			name:     "valid: empty line",
			content:  "\n",
			wantErrs: nil,
		},
		{
			name:     "valid: bold ending with colon",
			content:  "**Note:**\n",
			wantErrs: nil,
		},
		{
			name:     "valid: bold ending with semicolon",
			content:  "**Warning;**\n",
			wantErrs: nil,
		},
		// ── full-width / CJK punctuation ─────────────────────────────────────
		{
			name:     "valid: bold ending with ideographic full stop",
			content:  "**これは文章です。**\n",
			wantErrs: nil,
		},
		{
			name:     "valid: bold ending with Japanese comma",
			content:  "**ここで区切り、**\n",
			wantErrs: nil,
		},
		{
			name:     "valid: bold ending with full-width exclamation",
			content:  "**注意！**\n",
			wantErrs: nil,
		},
		{
			name:     "valid: bold ending with full-width question mark",
			content:  "**本当？**\n",
			wantErrs: nil,
		},
		{
			name:     "valid: bold ending with full-width colon",
			content:  "**補足：**\n",
			wantErrs: nil,
		},
		{
			name:    "invalid: Japanese heading without punctuation",
			content: "**はじめに**\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-emphasis-as-heading: emphasis used as heading, use ATX heading instead: **はじめに**"},
			},
		},
		{
			name:    "invalid: italic Japanese heading",
			content: "*概要*\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-emphasis-as-heading: emphasis used as heading, use ATX heading instead: *概要*"},
			},
		},

		// ── invalid cases ────────────────────────────────────────────────────
		{
			name:    "invalid: double-asterisk bold",
			content: "**Section Title**\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-emphasis-as-heading: emphasis used as heading, use ATX heading instead: **Section Title**"},
			},
		},
		{
			name:    "invalid: double-underscore bold",
			content: "__Section Title__\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-emphasis-as-heading: emphasis used as heading, use ATX heading instead: __Section Title__"},
			},
		},
		{
			name:    "invalid: single-asterisk italic",
			content: "*Section Title*\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-emphasis-as-heading: emphasis used as heading, use ATX heading instead: *Section Title*"},
			},
		},
		{
			name:    "invalid: single-underscore italic",
			content: "_Section Title_\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-emphasis-as-heading: emphasis used as heading, use ATX heading instead: _Section Title_"},
			},
		},
		{
			name:    "invalid: bold with leading whitespace",
			content: "  **Indented Bold**\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-emphasis-as-heading: emphasis used as heading, use ATX heading instead: **Indented Bold**"},
			},
		},
		{
			name:    "invalid: second line",
			content: "# Title\n\n**Subsection**\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "no-emphasis-as-heading: emphasis used as heading, use ATX heading instead: **Subsection**"},
			},
		},
		{
			name:    "invalid: offset shifts line numbers",
			content: "**Heading**\n",
			offset:  10,
			wantErrs: []LintError{
				{File: "test.md", Line: 11, Message: "no-emphasis-as-heading: emphasis used as heading, use ATX heading instead: **Heading**"},
			},
		},
		{
			name:    "invalid: multiple violations",
			content: "**First**\n\n__Second__\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-emphasis-as-heading: emphasis used as heading, use ATX heading instead: **First**"},
				{File: "test.md", Line: 3, Message: "no-emphasis-as-heading: emphasis used as heading, use ATX heading instead: __Second__"},
			},
		},
		{
			name:     "valid: empty file",
			content:  "",
			wantErrs: nil,
		},
		// ── inline code on same line ─────────────────────────────────────────
		{
			name:     "valid: emphasis plus inline code on same line",
			content:  "**Heading** `code`\n",
			wantErrs: nil,
		},
		{
			name:     "valid: entire line is inline code span",
			content:  "`**bold**`\n",
			wantErrs: nil,
		},
		// ── nested emphasis with trailing punctuation ─────────────────────────
		{
			name:     "valid: triple-asterisk bold-italic ending with colon",
			content:  "***Note:***\n",
			wantErrs: nil,
		},
		{
			name:     "valid: triple-asterisk bold-italic ending with period",
			content:  "***This is a sentence.***\n",
			wantErrs: nil,
		},
		{
			name:    "invalid: triple-asterisk bold-italic without punctuation",
			content: "***Section Title***\n",
			wantErrs: []LintError{
				{File: "test.md", Line: 1, Message: "no-emphasis-as-heading: emphasis used as heading, use ATX heading instead: ***Section Title***"},
			},
		},
		// ── delimiter inside span (not a heading) ────────────────────────────
		{
			name:     "valid: double-asterisk with nested delimiter",
			content:  "**bold**text**\n",
			wantErrs: nil,
		},
		{
			name:     "valid: double-underscore with nested delimiter",
			content:  "__under__score__\n",
			wantErrs: nil,
		},
		{
			name:     "valid: single-asterisk with nested delimiter",
			content:  "*ital*ic*\n",
			wantErrs: nil,
		},
		{
			name:     "valid: single-underscore with nested delimiter",
			content:  "_under_score_\n",
			wantErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(tt.content, "\n")
			got := CheckNoEmphasisAsHeading("test.md", lines, tt.offset)

			if len(got) != len(tt.wantErrs) {
				t.Fatalf("got %d errors, want %d\ngot:  %v\nwant: %v", len(got), len(tt.wantErrs), got, tt.wantErrs)
			}
			for i := range got {
				if got[i].File != tt.wantErrs[i].File {
					t.Errorf("[%d] file: got %q, want %q", i, got[i].File, tt.wantErrs[i].File)
				}
				if got[i].Line != tt.wantErrs[i].Line {
					t.Errorf("[%d] line: got %d, want %d", i, got[i].Line, tt.wantErrs[i].Line)
				}
				if got[i].Message != tt.wantErrs[i].Message {
					t.Errorf("[%d] message: got %q, want %q", i, got[i].Message, tt.wantErrs[i].Message)
				}
			}
		})
	}
}

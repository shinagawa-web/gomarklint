package preprocess

import (
	"strings"
	"testing"
)

// blanks returns a string of n spaces, matching how sanitizeInline blanks a
// region of length n. Used to build expected output without hand-counting.
func blanks(n int) string { return strings.Repeat(" ", n) }

func TestSanitizeInline(t *testing.T) {
	tests := []struct {
		name           string
		line           string
		startInComment bool
		wantSanitized  string
		wantEnded      bool
		wantFully      bool
	}{
		{
			name:          "plain prose unchanged",
			line:          "just some text",
			wantSanitized: "just some text",
		},
		{
			name:          "inline code span blanked, length preserved",
			line:          "see `code` here",
			wantSanitized: "see " + blanks(len("`code`")) + " here",
		},
		{
			name:          "multi-backtick span closes on equal run",
			line:          "a ``b`c`` d",
			wantSanitized: "a " + blanks(len("``b`c``")) + " d",
		},
		{
			name:          "unclosed backtick run is literal",
			line:          "a `b c",
			wantSanitized: "a `b c",
		},
		{
			name:          "inline comment blanked, prose kept",
			line:          "text <!-- note --> more",
			wantSanitized: "text " + blanks(len("<!-- note -->")) + " more",
		},
		{
			name:          "standalone comment is fully comment",
			line:          "<!-- only a comment -->",
			wantSanitized: blanks(len("<!-- only a comment -->")),
			wantFully:     true,
		},
		{
			name:          "unclosed comment carries state to next line",
			line:          "text <!-- start",
			wantSanitized: "text " + blanks(len("<!-- start")),
			wantEnded:     true,
		},
		{
			name:           "continuation line inside comment is fully comment",
			line:           "still inside the comment",
			startInComment: true,
			wantSanitized:  blanks(len("still inside the comment")),
			wantEnded:      true,
			wantFully:      true,
		},
		{
			name:           "comment closes mid-line, trailing prose kept",
			line:           "end --> visible",
			startInComment: true,
			wantSanitized:  blanks(len("end -->")) + " visible",
			wantEnded:      false,
		},
		{
			// Left-to-right: "<!--" inside a code span is code, not a comment,
			// so the line ends outside any comment.
			name:          "comment delimiter inside code span is not a comment",
			line:          "`<!--` text",
			wantSanitized: blanks(len("`<!--`")) + " text",
			wantEnded:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ended, fully := sanitizeInline(tt.line, tt.startInComment)
			if got != tt.wantSanitized {
				t.Errorf("sanitized = %q, want %q", got, tt.wantSanitized)
			}
			if len(got) != len(tt.line) {
				t.Errorf("length not preserved: got %d, want %d", len(got), len(tt.line))
			}
			if ended != tt.wantEnded {
				t.Errorf("endedInComment = %v, want %v", ended, tt.wantEnded)
			}
			if fully != tt.wantFully {
				t.Errorf("fullyComment = %v, want %v", fully, tt.wantFully)
			}
		})
	}
}

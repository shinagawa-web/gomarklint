package rule

import "testing"

// TestStripInlineCode covers the shared helper directly. The unclosed-backtick
// branch in particular is no longer exercised transitively by no-bare-urls,
// which migrated to the preprocess context (#337 Phase 2).
func TestStripInlineCode(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"no backticks", "plain text", "plain text"},
		{"single span blanked", "a `code` b", "a        b"},
		{"multi-backtick span blanked", "a ``c`d`` e", "a         e"},
		{
			// Unclosed run: the backticks have no matching closing run, so they
			// are emitted literally and the rest of the line is left intact.
			name: "unclosed backtick emitted literally",
			in:   "see `http://x.com here",
			want: "see `http://x.com here",
		},
		{
			// A longer unclosed run is also emitted verbatim.
			name: "unclosed double backtick emitted literally",
			in:   "``unterminated",
			want: "``unterminated",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripInlineCode(tt.in); got != tt.want {
				t.Errorf("stripInlineCode(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

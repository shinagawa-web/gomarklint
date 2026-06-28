package preprocess

import "testing"

func TestHTMLBlockStart(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		inParagraph bool
		want        int
	}{
		// Type 1: script/pre/style/textarea.
		{"type1 pre", "<pre>", false, 1},
		{"type1 script with attr", "<script type=\"text/js\">", false, 1},
		{"type1 style bare", "<style", false, 1},
		{"type1 case insensitive", "<SCRIPT>", false, 1},
		// Type 3: processing instruction.
		{"type3", "<?php echo 1; ?>", false, 3},
		// Type 4: declaration.
		{"type4 doctype", "<!DOCTYPE html>", false, 4},
		// Type 5: CDATA.
		{"type5 cdata", "<![CDATA[stuff]]>", false, 5},
		// Type 6: block-level tags. div is the load-bearing audit case.
		{"type6 div open", "<div>", false, 6},
		{"type6 div with attr", "<div class=\"x\">", false, 6},
		{"type6 div bare eol", "<div", false, 6},
		{"type6 closing div", "</div>", false, 6},
		{"type6 self-closing hr", "<hr/>", false, 6},
		{"type6 table", "<table>", false, 6},
		{"type6 can interrupt paragraph", "<div>", true, 6},
		// Type 7: a complete standalone tag that is not a type-1 tag.
		{"type7 custom open tag", "<custom-element>", false, 7},
		{"type7 custom closing tag", "</custom-element>", false, 7},
		{"type7 cannot interrupt paragraph", "<custom-element>", true, 0},
		// Non-starts.
		{"plain text", "not html", false, 0},
		{"inline tag mid-line is not type7", "see <span>x</span> here", false, 0},
		{"tag name cannot start with a digit", "<1numeric>", false, 0},
		{"comment is handled elsewhere", "<!-- comment -->", false, 0},
		{"angle bracket only", "<", false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := htmlBlockStart(tt.line, tt.inParagraph); got != tt.want {
				t.Errorf("htmlBlockStart(%q, inParagraph=%v) = %d, want %d",
					tt.line, tt.inParagraph, got, tt.want)
			}
		})
	}
}

func TestMatchType6(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"<div>", true},
		{"<hr/>", true},
		// A block tag name immediately followed by a character that is neither a
		// delimiter (space/tab/>) nor part of the tag name is not a type 6 start.
		{"<p=x>", false},
		{"<div@>", false},
		// Self-closing form must actually be "/>".
		{"<div/x>", false},
	}
	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			if got := matchType6(tt.line); got != tt.want {
				t.Errorf("matchType6(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestMatchType7(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"<custom-element>", true},
		{"</custom-element>", true},
		// Two tags on one line is not a single complete tag.
		{"<a><b>", false},
		// A type 1 tag name is excluded from type 7 (handled as type 1 instead).
		{"<pre>", false},
		{"<script>", false},
		// A closing tag carries no attributes.
		{"</div foo>", false},
	}
	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			if got := matchType7(tt.line); got != tt.want {
				t.Errorf("matchType7(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestHTMLBlockEndsOnLine(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		blockType int
		want      bool
	}{
		{"type1 closes on </pre>", "output</pre>", 1, true},
		{"type1 closes on </script>", "</script>", 1, true},
		{"type1 no close", "console.log(1)", 1, false},
		{"type3 closes on ?>", "done ?>", 3, true},
		{"type3 no close", "still going", 3, false},
		{"type4 closes on >", "html>", 4, true},
		{"type4 no close", "still declaring", 4, false},
		{"type5 closes on ]]>", "data]]>", 5, true},
		{"type5 no close", "more data", 5, false},
		// Types 6 and 7 end on a blank line, never via this function.
		{"type6 never ends here", "</div>", 6, false},
		{"type7 never ends here", "</custom-element>", 7, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := htmlBlockEndsOnLine(tt.line, tt.blockType); got != tt.want {
				t.Errorf("htmlBlockEndsOnLine(%q, %d) = %v, want %v",
					tt.line, tt.blockType, got, tt.want)
			}
		})
	}
}

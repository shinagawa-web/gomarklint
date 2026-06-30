package preprocess

import (
	"strings"
	"testing"
)

// flags is a compact view of the four context booleans for one line, used to
// assert Scan's output against the issue #337 audit matrix.
type flags struct {
	fenced, indented, htmlBlock, htmlComment bool
}

func flagsOf(c LineContext) flags {
	return flags{c.InFencedCode, c.InIndentedCode, c.InHTMLBlock, c.InHTMLComment}
}

// runScan splits a multi-line literal and runs Scan. A leading newline in the
// literal is dropped so test inputs can start on their own line.
func runScan(t *testing.T, doc string) []LineContext {
	t.Helper()
	doc = strings.TrimPrefix(doc, "\n")
	return Scan(strings.Split(doc, "\n"))
}

func TestScanContextFlags(t *testing.T) {
	tests := []struct {
		name string
		doc  string
		want []flags
	}{
		{
			// CommonMark §4.4 example 108: an indented code block cannot
			// interrupt a paragraph, so the indented line is a lazy
			// continuation of "Foo", not code.
			name: "indented line does not interrupt paragraph",
			doc: `
Foo
    bar`,
			want: []flags{{}, {}},
		},
		{
			// A four-space indented line after a blank line IS an indented
			// code block. Audit: heading rules / link-fragments must skip the
			// "# Fake" line instead of treating it as a real heading.
			name: "indented code block after blank line",
			doc: `
text

    # Fake heading`,
			want: []flags{{}, {}, {indented: true}},
		},
		{
			// Audit: a fenced code block opened at column >= 4 is indented code,
			// not a fence. The two detectors must not fight.
			name: "backtick fence indented four spaces is indented code",
			doc: `
text

    ` + "```" + `
    not a fence`,
			want: []flags{{}, {}, {indented: true}, {indented: true}},
		},
		{
			// Audit: a heading inside an HTML block (<div>) must not be parsed
			// as a heading. The block runs to a blank line (or EOF).
			name: "html block keeps inner heading in context",
			doc: `
<div>
# Not a heading
</div>`,
			want: []flags{{htmlBlock: true}, {htmlBlock: true}, {htmlBlock: true}},
		},
		{
			// A type-6 HTML block ends on a blank line; the blank line and what
			// follows are outside the block.
			name: "html block ends on blank line",
			doc: `
<div>
content
</div>

after`,
			want: []flags{
				{htmlBlock: true}, {htmlBlock: true}, {htmlBlock: true},
				{}, {},
			},
		},
		{
			// Audit: a multi-line HTML comment, including any "heading" inside
			// it, is comment context, not headings.
			name: "multi-line html comment",
			doc: `
<!--
# commented heading
-->
text`,
			want: []flags{
				{htmlComment: true}, {htmlComment: true}, {htmlComment: true},
				{},
			},
		},
		{
			// A multi-line comment can start mid-line after prose and close
			// mid-line before more prose. The opening and closing lines carry
			// real prose, so neither is a pure comment line; only the fully
			// commented middle line is flagged. Without this, downstream rules
			// would skip the trailing "visible prose".
			name: "multi-line comment opened and closed mid-line keeps prose",
			doc: `
text <!-- start
inside comment
end --> visible prose`,
			want: []flags{
				{}, {htmlComment: true}, {},
			},
		},
		{
			// Audit worst offender: empty-alt-text fires inside fenced code.
			// The image line must be flagged as fenced so the rule skips it.
			name: "fenced code block covers delimiters and content",
			doc: `
` + "```" + `
![](img.png)
` + "```",
			want: []flags{{fenced: true}, {fenced: true}, {fenced: true}},
		},
		{
			// Audit item 5: an unclosed fence must mark every line to EOF as
			// fenced, so downstream rules cannot mispair fences afterward.
			name: "unclosed fence flags all lines to EOF",
			doc: `
text

` + "```" + `
code
more code`,
			want: []flags{
				{}, {},
				{fenced: true}, {fenced: true}, {fenced: true},
			},
		},
		{
			// A type 1 HTML block (<pre>) spans multiple lines and ends only on
			// the line containing its matching close tag, not on a blank line.
			name: "multi-line type 1 html block ends on close tag",
			doc: `
<pre>
code
</pre>
after`,
			want: []flags{
				{htmlBlock: true}, {htmlBlock: true}, {htmlBlock: true},
				{},
			},
		},
		{
			// A type 4 HTML block (declaration) opens and ends on the same line
			// once its '>' delimiter is seen, so the following line is fresh.
			name: "single-line type 4 html block",
			doc: `
<!DOCTYPE html>
text`,
			want: []flags{{htmlBlock: true}, {}},
		},
		{
			// A backtick fence is not closed by a tilde fence of the same run
			// length; only a matching-character fence closes it.
			name: "tilde line does not close backtick fence",
			doc: `
` + "```" + `
~~~
still code
` + "```",
			want: []flags{
				{fenced: true}, {fenced: true}, {fenced: true}, {fenced: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runScan(t, tt.doc)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d lines, want %d:\n%#v", len(got), len(tt.want), got)
			}
			for i := range tt.want {
				if g := flagsOf(got[i]); g != tt.want[i] {
					t.Errorf("line %d (%q): flags = %+v, want %+v",
						i, got[i].Original, g, tt.want[i])
				}
			}
		})
	}
}

// TestScanSanitizedContract checks the inline-vs-block split of the Sanitized
// field described in the LineContext docs.
func TestScanSanitizedContract(t *testing.T) {
	t.Run("inline code in prose is blanked", func(t *testing.T) {
		ctx := runScan(t, "see `http://x.com` here")
		if strings.Contains(ctx[0].Sanitized, "http://x.com") {
			t.Errorf("inline code not blanked: %q", ctx[0].Sanitized)
		}
	})

	t.Run("fenced code content is left verbatim", func(t *testing.T) {
		doc := "```\n`backtick` text\n```"
		ctx := Scan(strings.Split(doc, "\n"))
		// Section B rules deliberately scan inside fenced code, so the content
		// line's Sanitized must equal its Original (no inline stripping).
		if ctx[1].Sanitized != ctx[1].Original {
			t.Errorf("fenced content was sanitized: got %q, want %q",
				ctx[1].Sanitized, ctx[1].Original)
		}
	})

	t.Run("standalone comment line is fully blanked", func(t *testing.T) {
		ctx := runScan(t, "<!-- a comment -->")
		if strings.TrimSpace(ctx[0].Sanitized) != "" {
			t.Errorf("comment line not blanked: %q", ctx[0].Sanitized)
		}
		if !ctx[0].InHTMLComment {
			t.Errorf("standalone comment not flagged InHTMLComment")
		}
	})

	t.Run("prose after a mid-line comment close is preserved", func(t *testing.T) {
		ctx := runScan(t, "text <!-- start\nend --> visible prose")
		last := ctx[len(ctx)-1]
		if last.InHTMLComment {
			t.Errorf("line with trailing prose flagged InHTMLComment: %q", last.Original)
		}
		if !strings.Contains(last.Sanitized, "visible prose") {
			t.Errorf("trailing prose not preserved in Sanitized: %q", last.Sanitized)
		}
	})

	t.Run("flags are mutually exclusive", func(t *testing.T) {
		doc := "para\n\n```\ncode\n```\n\n<div>\nx\n</div>\n\n    indented\n\n<!-- c -->"
		for i, c := range Scan(strings.Split(doc, "\n")) {
			n := 0
			for _, b := range []bool{c.InFencedCode, c.InIndentedCode, c.InHTMLBlock, c.InHTMLComment} {
				if b {
					n++
				}
			}
			if n > 1 {
				t.Errorf("line %d (%q) has %d context flags set, want <=1", i, c.Original, n)
			}
		}
	})

	t.Run("Original always preserved and Scan returns one ctx per line", func(t *testing.T) {
		lines := []string{"a", "  b `c`", "```", "d", "```"}
		ctx := Scan(lines)
		if len(ctx) != len(lines) {
			t.Fatalf("got %d contexts, want %d", len(ctx), len(lines))
		}
		for i := range lines {
			if ctx[i].Original != lines[i] {
				t.Errorf("line %d: Original = %q, want %q", i, ctx[i].Original, lines[i])
			}
		}
	})
}

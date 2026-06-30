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

func flagsOf(c *Context, i int) flags {
	return flags{c.InFencedCode(i), c.InIndentedCode(i), c.InHTMLBlock(i), c.InHTMLComment(i)}
}

// runScan splits a multi-line literal and runs Scan. A leading newline in the
// literal is dropped so test inputs can start on their own line.
func runScan(t *testing.T, doc string) *Context {
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
			if got.Len() != len(tt.want) {
				t.Fatalf("got %d lines, want %d", got.Len(), len(tt.want))
			}
			for i := range tt.want {
				if g := flagsOf(got, i); g != tt.want[i] {
					t.Errorf("line %d (%q): flags = %+v, want %+v",
						i, got.Line(i), g, tt.want[i])
				}
			}
		})
	}
}

// TestScanSanitizedContract checks the inline-vs-block split of the Sanitized
// view described in the Context docs.
func TestScanSanitizedContract(t *testing.T) {
	t.Run("inline code in prose is blanked", func(t *testing.T) {
		ctx := runScan(t, "see `http://x.com` here")
		if strings.Contains(ctx.Sanitized(0), "http://x.com") {
			t.Errorf("inline code not blanked: %q", ctx.Sanitized(0))
		}
	})

	t.Run("fenced code content is left verbatim", func(t *testing.T) {
		doc := "```\n`backtick` text\n```"
		ctx := Scan(strings.Split(doc, "\n"))
		// Section B rules deliberately scan inside fenced code, so the content
		// line's Sanitized must equal its Line (no inline stripping).
		if ctx.Sanitized(1) != ctx.Line(1) {
			t.Errorf("fenced content was sanitized: got %q, want %q",
				ctx.Sanitized(1), ctx.Line(1))
		}
	})

	t.Run("standalone comment line is fully blanked", func(t *testing.T) {
		ctx := runScan(t, "<!-- a comment -->")
		if strings.TrimSpace(ctx.Sanitized(0)) != "" {
			t.Errorf("comment line not blanked: %q", ctx.Sanitized(0))
		}
		if !ctx.InHTMLComment(0) {
			t.Errorf("standalone comment not flagged InHTMLComment")
		}
	})

	t.Run("prose after a mid-line comment close is preserved", func(t *testing.T) {
		ctx := runScan(t, "text <!-- start\nend --> visible prose")
		last := ctx.Len() - 1
		if ctx.InHTMLComment(last) {
			t.Errorf("line with trailing prose flagged InHTMLComment: %q", ctx.Line(last))
		}
		if !strings.Contains(ctx.Sanitized(last), "visible prose") {
			t.Errorf("trailing prose not preserved in Sanitized: %q", ctx.Sanitized(last))
		}
	})

	t.Run("flags are mutually exclusive", func(t *testing.T) {
		doc := "para\n\n```\ncode\n```\n\n<div>\nx\n</div>\n\n    indented\n\n<!-- c -->"
		ctx := Scan(strings.Split(doc, "\n"))
		for i := 0; i < ctx.Len(); i++ {
			n := 0
			for _, b := range []bool{ctx.InFencedCode(i), ctx.InIndentedCode(i), ctx.InHTMLBlock(i), ctx.InHTMLComment(i)} {
				if b {
					n++
				}
			}
			if n > 1 {
				t.Errorf("line %d (%q) has %d context flags set, want <=1", i, ctx.Line(i), n)
			}
		}
	})

	t.Run("Line preserves input and Scan covers every line", func(t *testing.T) {
		lines := []string{"a", "  b `c`", "```", "d", "```"}
		ctx := Scan(lines)
		if ctx.Len() != len(lines) {
			t.Fatalf("got %d lines, want %d", ctx.Len(), len(lines))
		}
		for i := range lines {
			if ctx.Line(i) != lines[i] {
				t.Errorf("line %d: Line = %q, want %q", i, ctx.Line(i), lines[i])
			}
		}
	})
}

// TestScanSanitizedIsSparse locks the compact-storage contract: the sanitized
// map holds an entry only for lines whose sanitized form differs from the
// original (those carrying an inline code span or comment). Verbatim lines —
// blank, prose without inline code, block code, HTML blocks — are absent and
// fall through to the original via Sanitized.
func TestScanSanitizedIsSparse(t *testing.T) {
	lines := []string{
		"plain prose line",        // verbatim
		"",                        // blank, verbatim
		"text with `code` span",   // differs -> stored
		"```",                     // fence open, verbatim
		"code `not blanked` here", // inside fence, verbatim
		"```",                     // fence close, verbatim
		"<!-- standalone -->",     // fully comment -> stored
	}
	ctx := Scan(lines)

	differs := map[int]bool{2: true, 6: true}
	for i := range lines {
		_, stored := ctx.sanitized[i]
		if stored != differs[i] {
			t.Errorf("line %d (%q): stored=%v, want %v", i, lines[i], stored, differs[i])
		}
		// Regardless of storage, Sanitized must round-trip verbatim lines.
		if !differs[i] && ctx.Sanitized(i) != lines[i] {
			t.Errorf("line %d (%q): Sanitized=%q, want verbatim", i, lines[i], ctx.Sanitized(i))
		}
	}
	if len(ctx.sanitized) != 2 {
		t.Errorf("sanitized map has %d entries, want 2 (only differing lines)", len(ctx.sanitized))
	}
}

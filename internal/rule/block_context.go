package rule

import "github.com/shinagawa-web/gomarklint/v3/internal/preprocess"

// inBlockContext reports whether line i sits in any code or HTML block context —
// fenced code, indented code, an HTML block, or an HTML comment. It is the full
// block-skip set used by rules that treat all such lines as non-content (heading,
// link, and structure rules).
//
// It is deliberately not a method on preprocess.Context: the preprocess package
// exposes the four contexts individually because some rules skip only a subset
// (e.g. the markdownlint divergences max-line-length and no-hard-tabs scan inside
// fenced code). Rules wanting a different subset call the individual predicates.
func inBlockContext(ctx *preprocess.Context, i int) bool {
	return ctx.InFencedCode(i) || ctx.InIndentedCode(i) || ctx.InHTMLBlock(i) || ctx.InHTMLComment(i)
}

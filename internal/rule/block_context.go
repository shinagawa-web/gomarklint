package rule

import "github.com/shinagawa-web/gomarklint/v3/internal/preprocess"

// Not a method on preprocess.Context: some rules (max-line-length, no-hard-tabs)
// skip only a subset of block contexts and call the individual predicates directly.
func inBlockContext(ctx *preprocess.Context, i int) bool {
	return ctx.InFencedCode(i) || ctx.InIndentedCode(i) || ctx.InHTMLBlock(i) || ctx.InHTMLComment(i)
}

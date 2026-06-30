package rule

import (
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

// CheckBlanksAroundFences flags fenced code blocks that are not preceded or
// followed by a blank line. Fences at the start or end of the file are exempt
// from the respective check. Fences inside indented code, HTML blocks, and HTML
// comments are not real fences and are ignored.
//
// A standalone single-line HTML comment (<!-- ... --> on its own line) is
// transparent: it is invisible in rendered output, so it does not satisfy or
// break the "preceded by a blank line" requirement. The preceding-blank check
// therefore looks past such lines. Multi-line comment blocks and inline comments
// (lines with visible text) are opaque.
func CheckBlanksAroundFences(filename string, ctx *preprocess.Context, offset int) []LintError {
	var errs []LintError

	for _, span := range ctx.FenceSpans() {
		// Preceded by a blank line? Skip past transparent standalone comment
		// lines to find the nearest visible line.
		j := span.Start - 1
		for j >= 0 && isTransparentComment(ctx, j) {
			j--
		}
		if j >= 0 && firstNonSpaceByte(ctx.Line(j)) != 0 {
			errs = append(errs, LintError{
				File:    filename,
				Line:    offset + span.Start + 1,
				Message: "blanks-around-fences: fenced code block must be preceded by a blank line",
			})
		}

		// Followed by a blank line? Only closed fences have a line after them to
		// check; an unclosed fence (End == -1) runs to EOF.
		if span.End >= 0 && span.End+1 < ctx.Len() && firstNonSpaceByte(ctx.Line(span.End+1)) != 0 {
			errs = append(errs, LintError{
				File:    filename,
				Line:    offset + span.End + 1,
				Message: "blanks-around-fences: fenced code block must be followed by a blank line",
			})
		}
	}

	return errs
}

// isTransparentComment reports whether line i is a standalone single-line HTML
// comment (the whole line is a comment, and it both opens and closes on that
// line). Such lines render to nothing, so blanks-around-fences looks through them
// when checking for a preceding blank line. Multi-line comment lines carry only
// one of the markers and are therefore opaque.
func isTransparentComment(ctx *preprocess.Context, i int) bool {
	if !ctx.InHTMLComment(i) {
		return false
	}
	line := ctx.Line(i)
	return strings.Contains(line, "<!--") && strings.Contains(line, "-->")
}

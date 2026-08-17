package rule

import (
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

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

func isTransparentComment(ctx *preprocess.Context, i int) bool {
	if !ctx.InHTMLComment(i) {
		return false
	}
	line := ctx.Line(i)
	return strings.Contains(line, "<!--") && strings.Contains(line, "-->")
}

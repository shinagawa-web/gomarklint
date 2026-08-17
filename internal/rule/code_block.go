package rule

import (
	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

func CheckUnclosedCodeBlocks(filename string, ctx *preprocess.Context, offset int) []LintError {
	var errs []LintError

	for _, span := range ctx.FenceSpans() {
		if span.End == -1 {
			errs = append(errs, LintError{
				File:    filename,
				Line:    span.Start + offset + 1,
				Message: "Unclosed code block",
			})
		}
	}

	return errs
}

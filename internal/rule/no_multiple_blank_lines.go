package rule

import (
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

func CheckNoMultipleBlankLines(filename string, ctx *preprocess.Context, offset int) []LintError {
	var errs []LintError
	consecutiveBlankCount := 0

	for i := 0; i < ctx.Len(); i++ {
		if inBlockContext(ctx, i) {
			consecutiveBlankCount = 0
			continue
		}
		if strings.TrimSpace(ctx.Line(i)) == "" {
			consecutiveBlankCount++
			if consecutiveBlankCount > 1 {
				errs = append(errs, LintError{
					File:    filename,
					Line:    i + 1 + offset,
					Message: "Multiple consecutive blank lines",
				})
			}
		} else {
			consecutiveBlankCount = 0
		}
	}

	return errs
}

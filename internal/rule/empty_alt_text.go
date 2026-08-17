package rule

import (
	"regexp"
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

var emptyAltTextRe = regexp.MustCompile(`!\[\s*\]\([^)]+\)`)

func CheckEmptyAltText(filename string, ctx *preprocess.Context, offset int) []LintError {
	var errs []LintError

	for i := 0; i < ctx.Len(); i++ {
		if inBlockContext(ctx, i) {
			continue
		}
		line := ctx.Sanitized(i)
		if !strings.Contains(line, "![") {
			continue
		}
		if emptyAltTextRe.MatchString(line) {
			errs = append(errs, LintError{
				File:    filename,
				Line:    i + 1 + offset,
				Message: "image with empty alt text",
			})
		}
	}

	return errs
}

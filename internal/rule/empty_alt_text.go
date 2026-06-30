package rule

import (
	"regexp"
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

var emptyAltTextRe = regexp.MustCompile(`!\[\s*\]\([^)]+\)`)

// CheckEmptyAltText reports lint errors for images with empty alt text.
//
// Images inside fenced code, indented code, HTML blocks, and HTML comments are
// skipped, and the inline-sanitized text is scanned so an ![](url) inside an
// inline code span or inline comment is ignored too. Before migrating to the
// preprocess context this rule did no context filtering at all (audit #337 worst
// offender), firing even inside fenced and inline code.
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

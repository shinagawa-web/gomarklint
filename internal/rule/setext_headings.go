package rule

import (
	"regexp"
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

var (
	setextUnderlineRegex  = regexp.MustCompile(`^ {0,3}(?:=+|-+)\s*$`)
	setextOtherBlockRegex = regexp.MustCompile(`^ {0,3}(?:[*+-]|\d+[.)]|>)\s*`)
)

func CheckNoSetextHeadings(filename string, ctx *preprocess.Context, offset int) []LintError {
	var errs []LintError
	isPrevLineEmpty := true
	isPrevLineOtherBlock := false
	isInLazyBlockquote := false

	for i := 0; i < ctx.Len(); i++ {
		line := ctx.Line(i)
		trimmed := strings.TrimSpace(line)

		isCurrentLineEmpty := trimmed == ""
		isCurrentLineOtherBlock := setextOtherBlockRegex.MatchString(line)

		if !inBlockContext(ctx, i) && setextUnderlineRegex.MatchString(line) &&
			!isPrevLineEmpty && !isPrevLineOtherBlock && !isInLazyBlockquote {
			errs = append(errs, LintError{
				File:    filename,
				Line:    i + 1 + offset,
				Message: "Setext heading found (prefer ATX style instead)",
			})
		}

		if isCurrentLineEmpty {
			isPrevLineEmpty = true
			isPrevLineOtherBlock = false
			isInLazyBlockquote = false
		} else if isCurrentLineOtherBlock {
			isPrevLineEmpty = false
			isPrevLineOtherBlock = true
			isInLazyBlockquote = strings.HasPrefix(trimmed, ">")
		} else {
			isPrevLineEmpty = false
			isPrevLineOtherBlock = false
		}
	}

	return errs
}

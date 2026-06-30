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

// CheckNoSetextHeadings ensures that headings of the "setext" style are never
// used; you should prefer ATX-type headings instead. It is better that one
// style is consistently used, and ATX headings are better because:
//
//  1. It's easier to remember
//  2. There is no risk of creating accidental <h2> elements by forgetting a
//     new line before a "----" horizontal rule.
//
// Parameters:
//   - filename: the name of the file being linted as a string
//   - ctx: the shared per-line context produced by preprocess.Scan
//   - offset: the line number offset due to frontmatter removal
//
// Returns:
//   - A slice of LintError containing the line number and description of each
//     detected issue.
//
// A setext underline inside fenced code, indented code, an HTML block, or an HTML
// comment is not a heading and is not reported.
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

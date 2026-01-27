package rule

import (
	"regexp"
	"strings"

	"github.com/shinagawa-web/gomarklint/internal/parser"
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
//   - content: the raw Markdown content as a string
//
// Returns:
//   - A slice of LintError containing the line number and description of each
//     detected issue.
func CheckNoSetextHeadings(filename, content string) []LintError {
	body, offset := parser.StripFrontmatter(content)
	lines := strings.Split(body, "\n")
	var errs []LintError

	// According to the CommonMark spec, a setext heading underline is
	// zero to 3 spaces, followed by any number of either the equals or
	// dash characters, optionally followed by whitespace.
	settextUnderlineRegex := regexp.MustCompile(`^ {0,3}(?:=+|-+)\s*$`)
	// A line is considered empty if it is of either no length or contains
	// only whitespace.
	emptyLineRegex := regexp.MustCompile(`^\s*$`)
	// List markers and blockquote markers
	otherBlockRegex := regexp.MustCompile(`^ {0,3}(?:[*+-]|\d+[.)]|>)\s*`)

	codeBlockRanges, _ := GetCodeBlockLineRanges(body)
	isPrevLineEmpty := true
	isPrevLineOtherBlock := false
	isInLazyBlockquote := false

	for i, line := range lines {
		isUnderline := settextUnderlineRegex.MatchString(line)
		isCurrentLineEmpty := emptyLineRegex.MatchString(line)
		isCurrentLineOtherBlock := otherBlockRegex.MatchString(line)
		if isUnderline && !isInCodeBlock(i+1, codeBlockRanges) {
			if !isPrevLineEmpty && !isPrevLineOtherBlock && !isInLazyBlockquote {
				errs = append(errs, LintError{
					File:    filename,
					Line:    i + 1 + offset,
					Message: "Setext heading found (prefer ATX style instead)",
				})
			}
		}
		if isCurrentLineEmpty {
			isPrevLineEmpty = true
			isPrevLineOtherBlock = false
			isInLazyBlockquote = false
		} else if isCurrentLineOtherBlock {
			isPrevLineEmpty = false
			isPrevLineOtherBlock = true
			isInLazyBlockquote = strings.HasPrefix(strings.TrimSpace(line), ">")
		} else {
			isPrevLineEmpty = false
			isPrevLineOtherBlock = false
		}
	}

	return errs
}

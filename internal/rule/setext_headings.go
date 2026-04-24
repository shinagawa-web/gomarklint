package rule

import (
	"regexp"
	"strings"
)

var (
	settextUnderlineRegex = regexp.MustCompile(`^ {0,3}(?:=+|-+)\s*$`)
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
//   - lines: the Markdown content split into lines (with frontmatter already removed)
//   - offset: the line number offset due to frontmatter removal
//
// Returns:
//   - A slice of LintError containing the line number and description of each
//     detected issue.
func CheckNoSetextHeadings(filename string, lines []string, offset int) []LintError {
	var errs []LintError
	inCodeBlock := false
	var fenceMarker string
	isPrevLineEmpty := true
	isPrevLineOtherBlock := false
	isInLazyBlockquote := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Maintain code-block state inline to avoid a separate O(n) pass and
		// the O(k) isInCodeBlock lookup per line.
		if inCodeBlock {
			if IsClosingFence(trimmed, fenceMarker) {
				inCodeBlock = false
				fenceMarker = ""
			}
		} else if marker := openingFenceMarker(trimmed); marker != "" {
			inCodeBlock = true
			fenceMarker = marker
		}

		isCurrentLineEmpty := trimmed == ""
		isCurrentLineOtherBlock := setextOtherBlockRegex.MatchString(line)

		if !inCodeBlock && settextUnderlineRegex.MatchString(line) &&
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

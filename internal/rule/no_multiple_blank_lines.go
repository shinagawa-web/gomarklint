package rule

import (
	"strings"
)

// CheckNoMultipleBlankLines checks whether the Markdown content contains
// multiple consecutive blank lines.
//
// This rule helps maintain consistency and readability by ensuring
// no more than one consecutive blank line appears in the document.
//
// It ignores multiple consecutive newlines in code blocks.
//
// Parameters:
//   - filename: the name of the file being checked (used in error reporting)
//   - lines: the Markdown content split into lines (with frontmatter already removed)
//   - offset: the line number offset due to frontmatter removal
//
// Returns:
//   - A slice of LintError with entries for each occurrence of multiple consecutive blank lines.
func CheckNoMultipleBlankLines(filename string, lines []string, offset int) []LintError {
	codeBlockRanges, _ := GetCodeBlockLineRanges(lines)

	var errs []LintError

	consecutiveBlankCount := 0
	for i, line := range lines {
		if !isInCodeBlock(i+1, codeBlockRanges) && strings.TrimSpace(line) == "" {
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

package rule

import (
	"strings"

	"github.com/shinagawa-web/gomarklint/internal/parser"
)

// CheckNoMultipleBlankLines checks whether the Markdown content contains
// multiple consecutive blank lines.
//
// This rule helps maintain consistency and readability by ensuring
// no more than one consecutive blank line appears in the document.
//
// It accounts for the presence of frontmatter, adjusting the reported
// line numbers accordingly, and ignores multiple consecutive newlines in code
// blocks.
//
// Parameters:
//   - filename: the name of the file being checked (used in error reporting)
//   - content: the raw Markdown content as a string
//
// Returns:
//   - A slice of LintError with entries for each occurrence of multiple consecutive blank lines.
func CheckNoMultipleBlankLines(filename, content string) []LintError {
	body, offset := parser.StripFrontmatter(content)

	codeBlockRanges, _ := GetCodeBlockLineRanges(body)

	var errs []LintError
	lines := strings.Split(body, "\n")

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

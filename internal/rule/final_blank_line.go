package rule

import (
	"strings"

	"github.com/shinagawa-web/gomarklint/internal/parser"
)

// CheckFinalBlankLine checks whether the Markdown content ends with a blank line.
// This is a common requirement in many Markdown style guides.
//
// It also accounts for the presence of frontmatter, adjusting the reported line number accordingly.
//
// Parameters:
//   - filename: the name of the file being checked (used in error reporting)
//   - content: the raw Markdown content as a string
//
// Returns:
//   - A slice of LintError with one entry if the final blank line is missing.
func CheckFinalBlankLine(filename, content string) []LintError {
	body, offset := parser.StripFrontmatter(content)

	var errs []LintError

	lines := strings.Split(body, "\n")
	if len(lines) < 2 || lines[len(lines)-1] != "" {
		errs = append(errs, LintError{
			File:    filename,
			Line:    len(lines) + offset,
			Message: "Missing final blank line",
		})
	}

	return errs
}

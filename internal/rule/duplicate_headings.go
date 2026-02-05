package rule

import (
	"fmt"
	"strings"
)

// CheckDuplicateHeadings detects duplicate headings in the given Markdown content.
// It treats headings as duplicates if their normalized text matches, regardless of level,
// case differences, or trailing spaces (including full-width spaces).
//
// This check helps enforce clear and non-redundant structure in documents,
// which improves readability and avoids confusion.
//
// Parameters:
//   - filename: the name of the file being checked (used in error reporting)
//   - lines: the Markdown content split into lines (with frontmatter already removed)
//   - offset: the line number offset due to frontmatter removal
//
// Returns:
//   - A slice of LintError entries for each detected duplicate heading (excluding the first occurrence).
func CheckDuplicateHeadings(filename string, lines []string, offset int) []LintError {
	var errs []LintError
	seen := map[string]int{}

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			heading := strings.TrimSpace(strings.TrimLeft(line, "#"))
			normalized := strings.ToLower(heading)

			if _, ok := seen[normalized]; ok {
				errs = append(errs, LintError{
					File:    filename,
					Line:    i + 1 + offset,
					Message: fmt.Sprintf("duplicate heading: %q", normalized),
				})
			} else {
				seen[normalized] = i + 1 + offset
			}
		}
	}

	return errs
}

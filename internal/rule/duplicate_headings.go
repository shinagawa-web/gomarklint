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
	seen := make(map[string]struct{}, len(lines)/10)
	inBlock := false
	fenceMarker := ""

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if inBlock {
			if IsClosingFence(trimmed, fenceMarker) {
				inBlock = false
				fenceMarker = ""
			}
			continue
		}

		if marker := openingFenceMarker(trimmed); marker != "" {
			inBlock = true
			fenceMarker = marker
			continue
		}

		if !isATXHeading(trimmed) {
			continue
		}

		heading := strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
		normalized := strings.ToLower(heading)

		if _, ok := seen[normalized]; ok {
			errs = append(errs, LintError{
				File:    filename,
				Line:    i + 1 + offset,
				Message: fmt.Sprintf("duplicate heading: %q", normalized),
			})
		} else {
			seen[normalized] = struct{}{}
		}
	}

	return errs
}

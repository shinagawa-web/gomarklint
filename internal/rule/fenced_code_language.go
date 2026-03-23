package rule

import "strings"

// CheckFencedCodeLanguage checks that all fenced code blocks specify a language identifier.
// Fenced code blocks opened with ``` or ~~~ without a language tag are flagged (MD040).
//
// Parameters:
//   - filename: the name of the file being checked (used in error reporting)
//   - lines: the Markdown content split into lines (with frontmatter already removed)
//   - offset: the line number offset due to frontmatter removal
//
// Returns:
//   - A slice of LintError, one per opening fence that is missing a language identifier.
func CheckFencedCodeLanguage(filename string, lines []string, offset int) []LintError {
	var errs []LintError
	inBlock := false
	fenceMarker := ""

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if inBlock {
			if strings.HasPrefix(trimmed, fenceMarker) && strings.TrimSpace(trimmed[len(fenceMarker):]) == "" {
				inBlock = false
				fenceMarker = ""
			}
			continue
		}

		marker := openingFenceMarker(trimmed)
		if marker == "" {
			continue
		}

		inBlock = true
		fenceMarker = marker

		if strings.TrimSpace(trimmed[len(marker):]) == "" {
			errs = append(errs, LintError{
				File:    filename,
				Line:    offset + i + 1,
				Message: "Fenced code block must have a language identifier",
			})
		}
	}

	return errs
}

// openingFenceMarker returns the fence marker ("```" or "~~~") if the line is an opening fence,
// or an empty string otherwise.
func openingFenceMarker(trimmed string) string {
	if strings.HasPrefix(trimmed, "```") {
		return "```"
	}
	if strings.HasPrefix(trimmed, "~~~") {
		return "~~~"
	}
	return ""
}

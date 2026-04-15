package rule

import "strings"

// CheckNoTrailingSpaces flags lines that end with one or more space or tab
// characters. Lines inside fenced code blocks are ignored.
func CheckNoTrailingSpaces(filename string, lines []string, offset int) []LintError {
	var errs []LintError
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

		// Strip trailing CR to handle CRLF line endings.
		stripped := strings.TrimRight(line, "\r")
		if len(stripped) > 0 && (stripped[len(stripped)-1] == ' ' || stripped[len(stripped)-1] == '\t') {
			errs = append(errs, LintError{
				File:    filename,
				Line:    offset + i + 1,
				Message: "no-trailing-spaces: trailing whitespace found",
			})
		}
	}

	return errs
}

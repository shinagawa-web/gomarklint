package rule

import "strings"

// CheckBlanksAroundFences flags fenced code blocks that are not preceded or
// followed by a blank line. Fences at the start or end of the file are exempt
// from the respective check. Fences inside HTML comment blocks are ignored.
func CheckBlanksAroundFences(filename string, lines []string, offset int) []LintError {
	var errs []LintError
	inBlock := false
	fenceMarker := ""
	inHTMLComment := false
	prevBlank := true

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		isBlank := trimmed == ""

		// Track HTML comment blocks (<!-- ... -->).
		// Fences inside comments are not real fenced code blocks.
		if !inHTMLComment {
			if strings.Contains(trimmed, "<!--") && !strings.Contains(trimmed, "-->") {
				inHTMLComment = true
				prevBlank = false
				continue
			}
		} else {
			if strings.Contains(trimmed, "-->") {
				inHTMLComment = false
			}
			prevBlank = false
			continue
		}

		if inBlock {
			if IsClosingFence(trimmed, fenceMarker) {
				inBlock = false
				fenceMarker = ""
				// closing fence: check the next line
				if i+1 < len(lines) && strings.TrimSpace(lines[i+1]) != "" {
					errs = append(errs, LintError{
						File:    filename,
						Line:    offset + i + 1,
						Message: "blanks-around-fences: fenced code block must be followed by a blank line",
					})
				}
			}
			prevBlank = false
			continue
		}

		if marker := openingFenceMarker(trimmed); marker != "" {
			inBlock = true
			fenceMarker = marker
			// opening fence: check the previous line
			if i > 0 && !prevBlank {
				errs = append(errs, LintError{
					File:    filename,
					Line:    offset + i + 1,
					Message: "blanks-around-fences: fenced code block must be preceded by a blank line",
				})
			}
			prevBlank = false
			continue
		}

		prevBlank = isBlank
	}

	return errs
}

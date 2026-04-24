package rule

import "strings"

// isATXHeading reports whether s is an ATX-style heading (levels 1–6).
func isATXHeading(s string) bool {
	level := 0
	for level < len(s) && s[level] == '#' {
		level++
	}
	if level == 0 || level > 6 {
		return false
	}
	return level == len(s) || s[level] == ' ' || s[level] == '\t'
}

// CheckBlanksAroundHeadings flags ATX-style headings that are not preceded or
// followed by a blank line. Headings at the start or end of the file are exempt
// from the respective check. Headings inside fenced code blocks are ignored.
func CheckBlanksAroundHeadings(filename string, lines []string, offset int) []LintError {
	var errs []LintError
	inBlock := false
	fenceMarker := ""
	// prevBlank tracks whether the previous line was blank so we avoid a
	// second TrimSpace call on lines[i-1] for every heading encountered.
	// Initialized to true so the first line of the file is exempt.
	prevBlank := true
	prevWasHeading := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		isBlank := trimmed == ""

		if inBlock {
			if IsClosingFence(trimmed, fenceMarker) {
				inBlock = false
				fenceMarker = ""
			}
			prevBlank = false
			prevWasHeading = false
			continue
		}

		if marker := openingFenceMarker(trimmed); marker != "" {
			inBlock = true
			fenceMarker = marker
			prevBlank = false
			prevWasHeading = false
			continue
		}

		// Check "followed by blank" for the previous heading on this iteration
		// so we avoid looking ahead with TrimSpace(lines[i+1]).
		if prevWasHeading && !isBlank {
			errs = append(errs, LintError{
				File:    filename,
				Line:    offset + i,
				Message: "blanks-around-headings: heading must be followed by a blank line",
			})
		}

		if isATXHeading(trimmed) {
			if i > 0 && !prevBlank {
				errs = append(errs, LintError{
					File:    filename,
					Line:    offset + i + 1,
					Message: "blanks-around-headings: heading must be preceded by a blank line",
				})
			}
			prevWasHeading = true
		} else {
			prevWasHeading = false
		}

		prevBlank = isBlank
	}

	return errs
}

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

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if inBlock {
			if IsClosingFence(trimmed, fenceMarker) {
				inBlock = false
				fenceMarker = ""
			}
			continue
		}

		marker := openingFenceMarker(trimmed)
		if marker != "" {
			inBlock = true
			fenceMarker = marker
			continue
		}

		if !isATXHeading(trimmed) {
			continue
		}

		if i > 0 && strings.TrimSpace(lines[i-1]) != "" {
			errs = append(errs, LintError{
				File:    filename,
				Line:    offset + i + 1,
				Message: "blanks-around-headings: heading must be preceded by a blank line",
			})
		}

		if i < len(lines)-1 && strings.TrimSpace(lines[i+1]) != "" {
			errs = append(errs, LintError{
				File:    filename,
				Line:    offset + i + 1,
				Message: "blanks-around-headings: heading must be followed by a blank line",
			})
		}
	}

	return errs
}

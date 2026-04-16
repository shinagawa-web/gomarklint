package rule

import "strings"

// isListItem reports whether line is a list item (unordered or ordered),
// allowing any amount of leading indentation.
func isListItem(line string) bool {
	s := strings.TrimLeft(line, " \t")
	if len(s) < 2 {
		return false
	}
	// Unordered: "- ", "* ", or "+ "
	if (s[0] == '-' || s[0] == '*' || s[0] == '+') && s[1] == ' ' {
		return true
	}
	// Ordered: one or more digits followed by '.' or ')' then a space
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i > 0 && i < len(s) && (s[i] == '.' || s[i] == ')') && i+1 < len(s) && s[i+1] == ' ' {
		return true
	}
	return false
}

// CheckBlanksAroundLists flags list blocks that are not preceded or followed by
// a blank line (MD032). The first item of a list block and the line immediately
// after the last item are checked. Lists at the start or end of the file are
// exempt from the respective check. List items inside fenced code blocks are
// ignored. Nested list items are treated as part of the same block and do not
// require additional blank lines between them and their parent.
func CheckBlanksAroundLists(filename string, lines []string, offset int) []LintError {
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

		if !isListItem(line) {
			continue
		}

		// Check "start of block": previous line is non-blank and not a list item.
		if i > 0 && strings.TrimSpace(lines[i-1]) != "" && !isListItem(lines[i-1]) {
			errs = append(errs, LintError{
				File:    filename,
				Line:    offset + i + 1,
				Message: "blanks-around-lists: list must be preceded by a blank line",
			})
		}

		// Check "end of block": next line is non-blank and not a list item.
		if i < len(lines)-1 && strings.TrimSpace(lines[i+1]) != "" && !isListItem(lines[i+1]) {
			errs = append(errs, LintError{
				File:    filename,
				Line:    offset + i + 2,
				Message: "blanks-around-lists: list must be followed by a blank line",
			})
		}
	}

	return errs
}

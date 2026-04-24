package rule

import "strings"

// isListItem reports whether line is a list item (unordered or ordered),
// allowing any amount of leading indentation. The marker must be followed by
// at least one space or tab, matching CommonMark's list item definition.
func isListItem(line string) bool {
	s := strings.TrimLeft(line, " \t")
	if len(s) < 2 {
		return false
	}
	return isUnorderedListItem(s) || isOrderedListItem(s)
}

// isUnorderedListItem reports whether s (already left-trimmed) starts with an
// unordered list marker ("- ", "* ", or "+ ").
func isUnorderedListItem(s string) bool {
	if s[0] != '-' && s[0] != '*' && s[0] != '+' {
		return false
	}
	return s[1] == ' ' || s[1] == '\t'
}

// isOrderedListItem reports whether s (already left-trimmed) starts with an
// ordered list marker: one or more digits followed by '.' or ')' then a space
// or tab.
func isOrderedListItem(s string) bool {
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == 0 || i >= len(s) {
		return false
	}
	if s[i] != '.' && s[i] != ')' {
		return false
	}
	if i+1 >= len(s) {
		return false
	}
	return s[i+1] == ' ' || s[i+1] == '\t'
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
	// prevBlank and prevWasListItem replace the TrimSpace look-behind on
	// lines[i-1] for every list item encountered.
	// prevBlank is initialized to true so the first line is exempt.
	prevBlank := true
	prevWasListItem := false
	prevLineNum := 0 // 1-indexed line number of the previous list item

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		isBlank := trimmed == ""
		isList := isListItem(line)

		// Check "end of block" before fence branches so a list item immediately
		// followed by a fence opener is still flagged (lesson from PR-4).
		if prevWasListItem && !isBlank && !isList {
			errs = append(errs, LintError{
				File:    filename,
				Line:    offset + prevLineNum + 1,
				Message: "blanks-around-lists: list must be followed by a blank line",
			})
		}

		if inBlock {
			if IsClosingFence(trimmed, fenceMarker) {
				inBlock = false
				fenceMarker = ""
			}
			prevBlank = false
			prevWasListItem = false
			continue
		}

		if marker := openingFenceMarker(trimmed); marker != "" {
			inBlock = true
			fenceMarker = marker
			prevBlank = false
			prevWasListItem = false
			continue
		}

		if isList {
			// Check "start of block": first item of a block not preceded by blank.
			if i > 0 && !prevBlank && !prevWasListItem {
				errs = append(errs, LintError{
					File:    filename,
					Line:    offset + i + 1,
					Message: "blanks-around-lists: list must be preceded by a blank line",
				})
			}
			prevWasListItem = true
			prevLineNum = i + 1
		} else {
			prevWasListItem = false
		}

		prevBlank = isBlank
	}

	return errs
}

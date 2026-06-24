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
		first := firstNonSpaceByte(line)
		isBlank := first == 0

		// HTML comment tracking only applies outside fenced code blocks;
		// `<!--`-like content inside a fenced block is just code and must not
		// interfere with detecting the closing fence.
		if !inBlock && (inHTMLComment || strings.IndexByte(line, '<') >= 0) {
			skip, stillInComment := stepHTMLComment(strings.TrimSpace(line), inHTMLComment)
			if skip {
				// Single-line HTML comments (<!-- ... --> on one line) are invisible
				// in rendered output; preserve prevBlank so they don't break a
				// blank-line chain established before the comment.
				// Multi-line comment lines (opening, body, or closing) are not
				// transparent and do reset the blank-line state.
				if inHTMLComment || stillInComment {
					prevBlank = false
				}
				inHTMLComment = stillInComment
				continue
			}
		}

		if inBlock {
			if first == fenceMarker[0] && IsClosingFence(strings.TrimSpace(line), fenceMarker) {
				inBlock = false
				fenceMarker = ""
				errs = appendMissingTrailingBlank(errs, filename, lines, i, offset)
			}
			prevBlank = false
			continue
		}

		if marker := openingFenceMarker(strings.TrimSpace(line)); marker != "" {
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

func appendMissingTrailingBlank(errs []LintError, filename string, lines []string, i, offset int) []LintError {
	if i+1 < len(lines) && firstNonSpaceByte(lines[i+1]) != 0 {
		return append(errs, LintError{
			File:    filename,
			Line:    offset + i + 1,
			Message: "blanks-around-fences: fenced code block must be followed by a blank line",
		})
	}
	return errs
}

// stepHTMLComment advances the HTML-comment state machine for one line.
// It returns (skip, inComment) where skip indicates the caller should
// `continue` past this line, and inComment is the updated state.
func stepHTMLComment(trimmed string, inComment bool) (bool, bool) {
	if inComment {
		if strings.Contains(trimmed, "-->") {
			return true, false
		}
		return true, true
	}
	if strings.Contains(trimmed, "<!--") {
		if strings.Contains(trimmed, "-->") {
			return true, false // single-line comment: <!-- ... -->
		}
		return true, true // opening of multi-line comment
	}
	return false, false
}

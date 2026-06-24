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
			skip, stillInComment, resetPrevBlank := stepHTMLComment(strings.TrimSpace(line), inHTMLComment)
			if skip {
				if resetPrevBlank {
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
// It returns (skip, newInComment, resetPrevBlank).
// skip: caller should continue past this line.
// newInComment: updated multi-line comment state.
// resetPrevBlank: true when the line contains visible content and prevBlank
// must be cleared; false for standalone single-line comments (<!-- ... -->)
// that are invisible in rendered output and should not break a blank-line chain.
func stepHTMLComment(trimmed string, inComment bool) (skip, newInComment, resetPrevBlank bool) {
	if inComment {
		if strings.Contains(trimmed, "-->") {
			return true, false, true
		}
		return true, true, true
	}
	if strings.Contains(trimmed, "<!--") {
		if strings.Contains(trimmed, "-->") {
			// Standalone single-line comment is transparent; mixed lines like
			// "text <!-- note -->" are not (HasPrefix distinguishes them).
			return true, false, !strings.HasPrefix(trimmed, "<!--")
		}
		return true, true, true
	}
	return false, false, false
}

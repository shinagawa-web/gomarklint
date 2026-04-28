package rule

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// atxHeadingText returns the visible text of an ATX heading with the opening
// '#' markers and optional closing '#' sequence removed.
// Returns ("", false) if the line is not a valid ATX heading.
func atxHeadingText(line string) (string, bool) {
	level := atxHeadingLevel(line)
	if level == 0 {
		return "", false
	}
	text := strings.TrimSpace(line[level:])
	// Strip optional closing ATX sequence: one or more '#' preceded by space or start.
	if n := len(text); n > 0 && text[n-1] == '#' {
		j := n - 1
		for j >= 0 && text[j] == '#' {
			j--
		}
		if j < 0 || text[j] == ' ' || text[j] == '\t' {
			text = strings.TrimRight(text[:j+1], " \t")
		}
	}
	return text, true
}

// CheckNoTrailingPunctuation flags ATX and setext headings whose visible text
// ends with a character contained in punctuation (MD026).
func CheckNoTrailingPunctuation(filename string, lines []string, offset int, punctuation string) []LintError {
	if punctuation == "" {
		return nil
	}

	var errs []LintError
	inBlock := false
	fenceMarker := ""
	prevText := ""       // trimmed text of the previous candidate setext-heading line
	prevIsBlock := false // true when the previous line was a block-level element

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if inBlock {
			if IsClosingFence(trimmed, fenceMarker) {
				inBlock = false
				fenceMarker = ""
				prevText = ""
				prevIsBlock = true
			}
			continue
		}

		if marker := openingFenceMarker(trimmed); marker != "" {
			inBlock = true
			fenceMarker = marker
			prevText = ""
			prevIsBlock = true
			continue
		}

		// ATX heading
		if text, ok := atxHeadingText(trimmed); ok {
			if r, ok := lastRuneInSet(text, punctuation); ok {
				errs = append(errs, LintError{
					File:    filename,
					Line:    i + 1 + offset,
					Message: fmt.Sprintf("no-trailing-punctuation: heading ends with %q", string(r)),
				})
			}
			prevText = ""
			prevIsBlock = true
			continue
		}

		// Setext heading underline — the heading text is on the previous line.
		// Use raw line (not trimmed) so 4-space-indented lines are not misidentified.
		if prevText != "" && !prevIsBlock && setextUnderlineRegex.MatchString(line) {
			if r, ok := lastRuneInSet(prevText, punctuation); ok {
				errs = append(errs, LintError{
					File:    filename,
					Line:    i + offset, // text line is lines[i-1]: 1-indexed i + offset
					Message: fmt.Sprintf("no-trailing-punctuation: heading ends with %q", string(r)),
				})
			}
			prevText = ""
			prevIsBlock = true
			continue
		}

		if trimmed == "" {
			prevText = ""
			prevIsBlock = false
		} else if setextOtherBlockRegex.MatchString(line) {
			prevText = ""
			prevIsBlock = true
		} else {
			prevText = trimmed
			prevIsBlock = false
		}
	}

	return errs
}

// lastRuneInSet reports whether the last Unicode rune of s is in set.
// Returns the rune and true when found, zero and false otherwise.
func lastRuneInSet(s, set string) (rune, bool) {
	if s == "" {
		return 0, false
	}
	r, _ := utf8.DecodeLastRuneInString(s)
	return r, strings.ContainsRune(set, r)
}

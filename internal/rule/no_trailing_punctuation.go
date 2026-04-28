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

// atxLineText returns the heading text if first == '#' and line is a valid ATX
// heading. Returns ("", false) otherwise.
func atxLineText(first byte, line string) (string, bool) {
	if first != '#' {
		return "", false
	}
	return atxHeadingText(strings.TrimSpace(line))
}

// openFenceMarkerIfPresent returns the fence marker when line opens a fenced
// code block, or "" otherwise. The first-byte guard avoids TrimSpace on most lines.
func openFenceMarkerIfPresent(first byte, line string) string {
	if first != '`' && first != '~' {
		return ""
	}
	return openingFenceMarker(strings.TrimSpace(line))
}

// setextHeadingText returns the trimmed heading text when line is a setext
// underline following a valid heading candidate on the previous line.
// Returns ("", false) when the conditions are not met.
func setextHeadingText(first byte, line, prevLine string, prevIsBlock bool) (string, bool) {
	if prevLine == "" || prevIsBlock || (first != '=' && first != '-') {
		return "", false
	}
	if !setextUnderlineRegex.MatchString(line) {
		return "", false
	}
	return strings.TrimSpace(prevLine), true
}

// isPossibleBlockMarker reports whether b can open a list item, ordered list,
// or block-quote line per CommonMark, making the line ineligible as setext text.
func isPossibleBlockMarker(b byte) bool {
	return b == '*' || b == '+' || b == '-' || b == '>' || (b >= '0' && b <= '9')
}

// noTPViolation builds a LintError for a heading that ends with rune r.
func noTPViolation(filename string, lineNum int, r rune) LintError {
	return LintError{
		File:    filename,
		Line:    lineNum,
		Message: fmt.Sprintf("no-trailing-punctuation: heading ends with %q", string(r)),
	}
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
	prevLine := ""       // raw previous non-blank non-block line (setext candidate)
	prevIsBlock := false // true when the previous line was a block-level element

	for i, line := range lines {
		first := firstNonSpaceByte(line)

		if first == 0 {
			prevLine = ""
			prevIsBlock = false
			continue
		}

		if inBlock {
			// Short-circuit: only trim+check when first byte matches the fence marker.
			if first == fenceMarker[0] && IsClosingFence(strings.TrimSpace(line), fenceMarker) {
				inBlock = false
				fenceMarker = ""
				prevLine = ""
				prevIsBlock = true
			}
			continue
		}

		if marker := openFenceMarkerIfPresent(first, line); marker != "" {
			inBlock = true
			fenceMarker = marker
			prevLine = ""
			prevIsBlock = true
			continue
		}

		if text, ok := atxLineText(first, line); ok {
			if r, ok := lastRuneInSet(text, punctuation); ok {
				errs = append(errs, noTPViolation(filename, i+1+offset, r))
			}
			prevLine = ""
			prevIsBlock = true
			continue
		}

		if text, ok := setextHeadingText(first, line, prevLine, prevIsBlock); ok {
			if r, ok := lastRuneInSet(text, punctuation); ok {
				errs = append(errs, noTPViolation(filename, i+offset, r))
			}
			prevLine = ""
			prevIsBlock = true
			continue
		}

		if isPossibleBlockMarker(first) && setextOtherBlockRegex.MatchString(line) {
			prevLine = ""
			prevIsBlock = true
			continue
		}

		prevLine = line
		prevIsBlock = false
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

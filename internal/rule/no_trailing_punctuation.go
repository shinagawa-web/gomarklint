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
	prevLine := ""       // raw previous non-blank non-block line (for setext detection)
	prevIsBlock := false // true when the previous line was a block-level element

	for i, line := range lines {
		first := firstNonSpaceByte(line)

		if first == 0 {
			// blank or whitespace-only line
			prevLine = ""
			prevIsBlock = false
			continue
		}

		if inBlock {
			if first == fenceMarker[0] {
				trimmed := strings.TrimSpace(line)
				if IsClosingFence(trimmed, fenceMarker) {
					inBlock = false
					fenceMarker = ""
					prevLine = ""
					prevIsBlock = true
				}
			}
			continue
		}

		// Opening fence — only lines starting with '`' or '~' can open a fence.
		if first == '`' || first == '~' {
			trimmed := strings.TrimSpace(line)
			if marker := openingFenceMarker(trimmed); marker != "" {
				inBlock = true
				fenceMarker = marker
				prevLine = ""
				prevIsBlock = true
				continue
			}
		}

		// ATX heading — only lines starting with '#'.
		if first == '#' {
			trimmed := strings.TrimSpace(line)
			if text, ok := atxHeadingText(trimmed); ok {
				if r, ok := lastRuneInSet(text, punctuation); ok {
					errs = append(errs, LintError{
						File:    filename,
						Line:    i + 1 + offset,
						Message: fmt.Sprintf("no-trailing-punctuation: heading ends with %q", string(r)),
					})
				}
				prevLine = ""
				prevIsBlock = true
				continue
			}
		}

		// Setext heading underline — only '=' or '-' lines can be underlines.
		// Use raw line (not trimmed) to honour the CommonMark 4-space indented-code rule.
		if prevLine != "" && !prevIsBlock && (first == '=' || first == '-') {
			if setextUnderlineRegex.MatchString(line) {
				prevText := strings.TrimSpace(prevLine)
				if r, ok := lastRuneInSet(prevText, punctuation); ok {
					errs = append(errs, LintError{
						File:    filename,
						Line:    i + offset, // text line is lines[i-1]: 1-indexed i + offset
						Message: fmt.Sprintf("no-trailing-punctuation: heading ends with %q", string(r)),
					})
				}
				prevLine = ""
				prevIsBlock = true
				continue
			}
		}

		// Detect other block-level elements (lists, blockquotes) so that the line
		// following them is not mistakenly treated as setext heading text.
		// Gate behind a first-byte check to avoid regex calls on paragraph lines.
		if first == '*' || first == '+' || first == '-' || first == '>' ||
			(first >= '0' && first <= '9') {
			if setextOtherBlockRegex.MatchString(line) {
				prevLine = ""
				prevIsBlock = true
				continue
			}
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

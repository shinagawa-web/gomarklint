package rule

import (
	"strings"
)

// CheckConsistentEmphasisStyle flags emphasis spans that use a different marker
// than expected. style must be "consistent", "asterisk", or "underscore".
//
// In "consistent" mode the first emphasis character found in the document sets
// the expected style; every subsequent span using a different character is
// flagged. In "asterisk"/"underscore" mode every span using the wrong character
// is flagged. Underscores inside words (e.g. snake_case) are not treated as
// emphasis. Content inside fenced code blocks and inline code spans is ignored.
func CheckConsistentEmphasisStyle(filename string, lines []string, offset int, style string) []LintError {
	var errs []LintError
	inBlock := false
	fenceMarker := ""
	var expectedCh byte // 0 until first emphasis seen (consistent mode)

	for i, line := range lines {
		first := firstNonSpaceByte(line)

		if inBlock {
			if first == fenceMarker[0] && IsClosingFence(strings.TrimSpace(line), fenceMarker) {
				inBlock = false
				fenceMarker = ""
			}
			continue
		}

		if first == '`' || first == '~' {
			if marker := openingFenceMarker(strings.TrimSpace(line)); marker != "" {
				inBlock = true
				fenceMarker = marker
				continue
			}
		}

		if !strings.ContainsAny(line, "*_") {
			continue
		}

		scanned := line
		if strings.ContainsRune(scanned, '`') {
			scanned = stripInlineCode(scanned)
		}

		checkEmphasisLine(scanned, filename, offset+i+1, style, &expectedCh, &errs)
	}

	return errs
}

// checkEmphasisLine scans s for emphasis spans and appends any style violations
// to errs. Only valid spans (opener + matching closer) are counted, so closing
// delimiters followed by punctuation are never double-counted. The function is
// allocation-free: no intermediate slice is created.
func checkEmphasisLine(s string, filename string, lineNum int, style string, expectedCh *byte, errs *[]LintError) {
	i := 0
	for i < len(s) {
		ch := s[i]

		if ch == '\\' {
			i += 2
			continue
		}

		if ch != '*' && ch != '_' {
			i++
			continue
		}

		runLen := 1
		for i+runLen < len(s) && s[i+runLen] == ch {
			runLen++
		}

		// Runs of 3+ are combinations (e.g. ***), not simple emphasis.
		if runLen > 2 {
			i += runLen
			continue
		}

		afterRun := i + runLen

		// Left-flanking: must be followed by a non-whitespace character.
		if !isEmphLeftFlanking(s, afterRun) {
			i += runLen
			continue
		}

		// Underscores flanked by word characters on both sides are mid-word.
		if ch == '_' && isEmphMidWord(s, i, afterRun) {
			i += runLen
			continue
		}

		// Require a matching closer so that closing delimiters (e.g. the `_`
		// in `_italic_.`) are never counted as openers.
		closerPos := findEmphCloser(s, afterRun, ch, runLen)
		if closerPos == -1 {
			i += runLen
			continue
		}

		if err := checkEmphasisStyle(filename, lineNum, ch, style, expectedCh); err != nil {
			*errs = append(*errs, *err)
		}
		i = closerPos + runLen // advance past the entire span
	}
}

// findEmphCloser returns the start position of the first right-flanking run of
// ch with exactly runLen characters at or after start. Returns -1 if not found.
func findEmphCloser(s string, start int, ch byte, runLen int) int {
	j := start
	for j < len(s) {
		if s[j] != ch {
			j++
			continue
		}
		closeLen := 1
		for j+closeLen < len(s) && s[j+closeLen] == ch {
			closeLen++
		}
		// Right-flanking: same length as opener and preceded by non-whitespace.
		if closeLen == runLen && j > 0 && s[j-1] != ' ' && s[j-1] != '\t' {
			return j
		}
		j += closeLen
	}
	return -1
}

// isEmphLeftFlanking reports whether afterRun is a valid left-flanking position
// (followed by a non-whitespace character).
func isEmphLeftFlanking(s string, afterRun int) bool {
	return afterRun < len(s) && s[afterRun] != ' ' && s[afterRun] != '\t'
}

// isEmphMidWord reports whether the delimiter run starting at i is a mid-word
// underscore (flanked by word characters on both sides).
func isEmphMidWord(s string, i, afterRun int) bool {
	return i > 0 && isEmphWordChar(s[i-1]) && isEmphWordChar(s[afterRun])
}

// isEmphWordChar reports whether b is a word character for emphasis purposes.
func isEmphWordChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

// checkEmphasisStyle validates ch against the configured style and updates
// expectedCh in consistent mode. Returns a LintError if the emphasis character
// does not match, nil otherwise.
func checkEmphasisStyle(filename string, line int, ch byte, style string, expectedCh *byte) *LintError {
	switch style {
	case "consistent":
		if *expectedCh == 0 {
			*expectedCh = ch
			return nil
		}
		if ch != *expectedCh {
			return &LintError{
				File:    filename,
				Line:    line,
				Message: "consistent-emphasis-style: expected " + emphCharName(*expectedCh) + " emphasis, got " + emphCharName(ch) + " emphasis",
			}
		}
	case "asterisk":
		if ch != '*' {
			return &LintError{
				File:    filename,
				Line:    line,
				Message: "consistent-emphasis-style: expected asterisk emphasis, got underscore emphasis",
			}
		}
	case "underscore":
		if ch != '_' {
			return &LintError{
				File:    filename,
				Line:    line,
				Message: "consistent-emphasis-style: expected underscore emphasis, got asterisk emphasis",
			}
		}
	}
	return nil
}

func emphCharName(ch byte) string {
	if ch == '*' {
		return "asterisk"
	}
	return "underscore"
}

package rule

import (
	"strings"
)

// CheckConsistentEmphasisStyle flags emphasis spans that use a different marker
// than expected. style must be "consistent", "asterisk", or "underscore";
// any other value falls back to "consistent".
//
// In "consistent" mode the first emphasis character found in the document sets
// the expected style; every subsequent span using a different character is
// flagged. In "asterisk"/"underscore" mode every span using the wrong character
// is flagged. Underscores inside words (e.g. snake_case) are not treated as
// emphasis. Content inside fenced code blocks and inline code spans is ignored.
func CheckConsistentEmphasisStyle(filename string, lines []string, offset int, style string) []LintError {
	switch style {
	case "consistent", "asterisk", "underscore":
	default:
		style = "consistent"
	}

	var errs []LintError
	inBlock := false
	fenceMarker := ""
	var expectedCh byte // 0 until first emphasis seen (consistent mode)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if inBlock {
			if IsClosingFence(trimmed, fenceMarker) {
				inBlock = false
				fenceMarker = ""
			}
			continue
		}

		if marker := openingFenceMarker(trimmed); marker != "" {
			inBlock = true
			fenceMarker = marker
			continue
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

// checkEmphasisLine scans s for emphasis openers and appends any style
// violations to errs. It is allocation-free: no intermediate slice is created.
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

		if err := checkEmphasisStyle(filename, lineNum, ch, style, expectedCh); err != nil {
			*errs = append(*errs, *err)
		}
		i += runLen
	}
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

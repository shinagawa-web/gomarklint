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
	var expectedEmphCh byte   // for runLen == 1 (emphasis)
	var expectedStrongCh byte // for runLen == 2 (strong)

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
		if strings.Contains(scanned, "](") {
			scanned = stripLinkURLs(scanned)
		}

		checkEmphasisLine(scanned, filename, offset+i+1, style, &expectedEmphCh, &expectedStrongCh, &errs)
	}

	return errs
}

// checkEmphasisLine scans s for emphasis spans and appends any style violations
// to errs. Only valid spans (opener + matching closer) are counted, so closing
// delimiters followed by punctuation are never double-counted. The function is
// allocation-free: no intermediate slice is created.
//
// expectedEmphCh tracks the expected marker for single-delimiter spans (emphasis);
// expectedStrongCh tracks it for double-delimiter spans (strong). They are kept
// separate so that *italic* and __strong__ can coexist without a violation.
func checkEmphasisLine(s string, filename string, lineNum int, style string, expectedEmphCh *byte, expectedStrongCh *byte, errs *[]LintError) {
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

		expected := expectedEmphCh
		kind := "emphasis"
		if runLen == 2 {
			expected = expectedStrongCh
			kind = "strong"
		}
		if err := checkEmphasisStyle(filename, lineNum, ch, style, expected, kind); err != nil {
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
// expectedCh in consistent mode. kind is "emphasis" for single-delimiter spans
// and "strong" for double-delimiter spans. Returns a LintError if the marker
// character does not match, nil otherwise.
func checkEmphasisStyle(filename string, line int, ch byte, style string, expectedCh *byte, kind string) *LintError {
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
				Message: "consistent-emphasis-style: expected " + emphCharName(*expectedCh) + " " + kind + ", got " + emphCharName(ch) + " " + kind,
			}
		}
	case "asterisk":
		if ch != '*' {
			return &LintError{
				File:    filename,
				Line:    line,
				Message: "consistent-emphasis-style: expected asterisk " + kind + ", got underscore " + kind,
			}
		}
	case "underscore":
		if ch != '_' {
			return &LintError{
				File:    filename,
				Line:    line,
				Message: "consistent-emphasis-style: expected underscore " + kind + ", got asterisk " + kind,
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

// stripLinkURLs replaces the destination and title of inline Markdown links
// with spaces so that underscores or asterisks in URLs are not mistaken for
// emphasis markers. The link text between [ and ] is preserved.
func stripLinkURLs(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		if s[i] != ']' || i+1 >= len(s) || s[i+1] != '(' || !hasPrecedingBracket(s, i) {
			b.WriteByte(s[i])
			i++
			continue
		}
		b.WriteByte(']')
		b.WriteByte('(')
		i += 2
		i = consumeLinkDest(&b, s, i)
		i = consumeLinkTitle(&b, s, i)
		if i < len(s) && s[i] == ')' {
			b.WriteByte(')')
			i++
		}
	}
	return b.String()
}

// hasPrecedingBracket reports whether there is an unescaped '[' before position i in s.
func hasPrecedingBracket(s string, i int) bool {
	for j := i - 1; j >= 0; j-- {
		if s[j] == '[' {
			return true
		}
	}
	return false
}

// consumeLinkDest blanks out the link destination starting at i and returns
// the position after it. Handles both angle-bracket form (<dest>) and the
// regular balanced-paren form.
func consumeLinkDest(b *strings.Builder, s string, i int) int {
	if i >= len(s) {
		return i
	}
	if s[i] == '<' {
		b.WriteByte('<')
		i++
		for i < len(s) && s[i] != '>' {
			b.WriteByte(' ')
			i++
		}
		if i < len(s) {
			b.WriteByte('>')
			i++
		}
		return i
	}
	depth := 0
	for i < len(s) {
		c := s[i]
		switch {
		case c == '(':
			depth++
			b.WriteByte(' ')
			i++
		case c == ')' && depth > 0:
			depth--
			b.WriteByte(' ')
			i++
		case c == ')' || c == ' ' || c == '\t':
			return i
		default:
			b.WriteByte(' ')
			i++
		}
	}
	return i
}

// consumeLinkTitle blanks out optional whitespace and the link title (if any)
// starting at i and returns the position after it.
func consumeLinkTitle(b *strings.Builder, s string, i int) int {
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		b.WriteByte(' ')
		i++
	}
	if i >= len(s) {
		return i
	}
	var closer byte
	switch s[i] {
	case '"':
		closer = '"'
	case '\'':
		closer = '\''
	case '(':
		closer = ')'
	default:
		return i
	}
	b.WriteByte(s[i])
	i++
	for i < len(s) && s[i] != closer {
		b.WriteByte(' ')
		i++
	}
	if i < len(s) {
		b.WriteByte(s[i])
		i++
	}
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		b.WriteByte(' ')
		i++
	}
	return i
}

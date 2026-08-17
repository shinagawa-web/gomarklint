package rule

import (
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

func CheckConsistentEmphasisStyle(filename string, ctx *preprocess.Context, offset int, style string) []LintError {
	var errs []LintError
	var expectedEmphCh byte   // for runLen == 1 (emphasis)
	var expectedStrongCh byte // for runLen == 2 (strong)

	for i := 0; i < ctx.Len(); i++ {
		if inBlockContext(ctx, i) {
			continue
		}

		// Sanitized blanks inline code spans and inline HTML comments, so
		// emphasis markers living inside them are not counted.
		scanned := ctx.Sanitized(i)
		if !strings.ContainsAny(scanned, "*_") {
			continue
		}
		if strings.Contains(scanned, "](") {
			scanned = stripLinkURLs(scanned)
		}

		checkEmphasisLine(scanned, filename, offset+i+1, style, &expectedEmphCh, &expectedStrongCh, &errs)
	}

	return errs
}

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

func isEmphLeftFlanking(s string, afterRun int) bool {
	return afterRun < len(s) && s[afterRun] != ' ' && s[afterRun] != '\t'
}

func isEmphMidWord(s string, i, afterRun int) bool {
	return i > 0 && isEmphWordChar(s[i-1]) && isEmphWordChar(s[afterRun])
}

func isEmphWordChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

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

func hasPrecedingBracket(s string, i int) bool {
	for j := i - 1; j >= 0; j-- {
		if s[j] == '[' {
			return true
		}
	}
	return false
}

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
